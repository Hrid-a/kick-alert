package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Hrid-a/kick-alert/internal/database"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

type Product struct {
	Slug          string    `json:"slug"`
	Name          string    `json:"name"`
	Sku           string    `json:"sku"`
	ExternalID    string    `json:"cloudProductId"`
	Category      string    `json:"category"`
	URL           string    `json:"url"`
	ImageURL      string    `json:"image_url"`
	CurrentPrice  string    `json:"current_price"`
	Currency      string    `json:"currency"`
	InStock       bool      `json:"in_stock"`
	LastScrapedAt time.Time `json:"last_scraped_at"`
}

func (app *application) addProductHandler(c *gin.Context) {
	userId := c.MustGet("userId").(uuid.UUID)

	var input struct {
		Url string `json:"product_url" binding:"required"`
	}

	if err := c.BindJSON(&input); err != nil {
		app.errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	limited, err := app.watchlistLimitReached(c, userId)
	if err != nil {
		app.serverErrorResponse(c, err)
		return
	}
	if limited {
		app.errorResponse(c, http.StatusForbidden, "watchlist limit reached: upgrade to pro for unlimited slots")
		return
	}

	bgCtx := context.Background()
	productURL := input.Url

	app.wg.Go(func() {
		defer func() {
			if pv := recover(); pv != nil {
				app.logger.Error(fmt.Sprintf("%v", pv))
			}
		}()

		item, err := app.scrapeNikeProduct(bgCtx, productURL)
		if err != nil {
			app.logger.Error("addProduct: scrape failed", "error", err.Error())
			return
		}

		// Reuse existing product row if we've already scraped this external_id.
		existing, err := app.db.GetProductByExternalID(bgCtx, item.ID)
		var productID uuid.UUID
		if err == nil {
			productID = existing.ID
		} else {
			slug := slugFromNikeURL(item.URL)
			if slug == "" {
				slug = strings.ToLower(strings.ReplaceAll(item.StyleColor, " ", "-"))
			}

			product, err := app.db.InsertProduct(bgCtx, database.InsertProductParams{
				Slug:          slug,
				Name:          item.Title,
				Sku:           item.StyleColor,
				ExternalID:    item.ID,
				Category:      "FOOTWEAR",
				Url:           item.URL,
				ImageUrl:      "",
				CurrentPrice:  fmt.Sprintf("%.2f", item.Price.Current),
				Currency:      item.Price.Currency,
				InStock:       item.Status == "BUYABLE_BUY",
				LastScrapedAt: time.Now(),
			})
			if err != nil {
				app.logger.Error("addProduct: insert product failed", "error", err.Error())
				return
			}
			productID = product.ID
		}

		_, err = app.db.InsertWatchlistEntry(bgCtx, database.InsertWatchlistEntryParams{
			UserID:       userId,
			ProductID:    productID,
			AlertSale:    true,
			AlertRestock: true,
		})
		if err != nil {
			var pqErr *pq.Error
			if !(errors.As(err, &pqErr) && pqErr.Code == "23505") {
				app.logger.Error("addProduct: insert watchlist entry failed", "error", err.Error())
			}
		}
	})

	c.Writer.WriteHeader(http.StatusAccepted)
}

func (app *application) getPriceHistoryHandler(c *gin.Context) {
	productId, err := uuid.Parse(c.Param("id"))
	if err != nil {
		app.errorResponse(c, http.StatusBadRequest, "invalid product id")
		return
	}

	limit := int32(50)
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 200 {
			limit = int32(parsed)
		}
	}

	history, err := app.db.GetPriceHistoryByProduct(c.Request.Context(), database.GetPriceHistoryByProductParams{
		ProductID: productId,
		Limit:     limit,
	})
	if err != nil {
		app.serverErrorResponse(c, err)
		return
	}

	type pricePoint struct {
		Price     string    `json:"price"`
		InStock   bool      `json:"in_stock"`
		ScrapedAt time.Time `json:"scraped_at"`
	}

	result := make([]pricePoint, len(history))
	for i, h := range history {
		result[i] = pricePoint{
			Price:     h.Price,
			InStock:   h.InStock,
			ScrapedAt: h.ScrapedAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{"price_history": result})
}

func (app *application) getProductHandler(c *gin.Context) {

	type response struct {
		Product Product `json:"product"`
	}

	id := c.Param("id")

	productId, err := uuid.Parse(id)

	if err != nil {
		app.errorResponse(c, http.StatusBadRequest, "couldn't proccess your request")
		return
	}

	product, err := app.db.GetProductById(c.Request.Context(), productId)

	if err != nil {

		switch {
		case errors.Is(err, sql.ErrNoRows):
			app.errorResponse(c, http.StatusNotFound, fmt.Sprintf("product with id %v is not found", id))
		default:
			app.serverErrorResponse(c, err)
		}
		return
	}

	c.JSON(http.StatusOK, response{
		Product: Product{
			Slug:          product.Slug,
			Name:          product.Name,
			Sku:           product.Sku,
			ExternalID:    product.ExternalID,
			Category:      product.Category,
			URL:           product.Url,
			ImageURL:      product.ImageUrl,
			CurrentPrice:  product.CurrentPrice,
			Currency:      product.Currency,
			InStock:       product.InStock,
			LastScrapedAt: product.LastScrapedAt,
		},
	})
}

func (app *application) getProductsHandler(c *gin.Context) {

	type response struct {
		Products []Product `json:"products"`
	}

	limit := 20
	offset := 0

	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	if page := c.Query("page"); page != "" {
		if parsed, err := strconv.Atoi(page); err == nil && parsed > 1 {
			offset = (parsed - 1) * limit
		}
	}

	params := database.SearchProductsParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	}

	if q := c.Query("q"); q != "" {
		params.Column1 = q
	}
	if cat := c.Query("category"); cat != "" {
		params.Column2 = cat
	}
	if s := c.Query("in_stock"); s != "" {
		params.Column3 = s
	}
	if min := c.Query("min_price"); min != "" {
		params.Column4 = min
	}
	if max := c.Query("max_price"); max != "" {
		params.Column5 = max
	}

	products, err := app.db.SearchProducts(c.Request.Context(), params)
	if err != nil {
		app.serverErrorResponse(c, err)
		return
	}

	result := make([]Product, len(products))
	for i, p := range products {
		result[i] = Product{
			Slug:          p.Slug,
			Name:          p.Name,
			Sku:           p.Sku,
			ExternalID:    p.ExternalID,
			Category:      p.Category,
			URL:           p.Url,
			ImageURL:      p.ImageUrl,
			CurrentPrice:  p.CurrentPrice,
			Currency:      p.Currency,
			InStock:       p.InStock,
			LastScrapedAt: p.LastScrapedAt,
		}
	}

	c.JSON(http.StatusOK, response{Products: result})
}
