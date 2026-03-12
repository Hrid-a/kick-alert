package main

import (
	"database/sql"
	"errors"
	"net/http"
	"time"

	"github.com/Hrid-a/kick-alert/internal/database"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

const freeTierLimit = 5

// watchlistLimitReached returns true if the user has hit their tier's watchlist cap.
func (app *application) watchlistLimitReached(c *gin.Context, userId uuid.UUID) (bool, error) {
	user, err := app.db.GetUserById(c.Request.Context(), userId)
	if err != nil {
		return false, err
	}
	if user.Tier.String == "pro" {
		return false, nil
	}
	count, err := app.db.CountUserWatchlist(c.Request.Context(), userId)
	if err != nil {
		return false, err
	}
	return count >= freeTierLimit, nil
}

type watchlistEntry struct {
	ID           uuid.UUID `json:"id"`
	ProductID    uuid.UUID `json:"product_id"`
	AlertSale    bool      `json:"alert_sale"`
	AlertRestock bool      `json:"alert_restock"`
	CreatedAt    time.Time `json:"created_at"`
}

type watchlistEntryWithProduct struct {
	ID                  uuid.UUID `json:"id"`
	ProductID           uuid.UUID `json:"product_id"`
	ProductName         string    `json:"product_name"`
	ProductSlug         string    `json:"product_slug"`
	ProductImageURL     string    `json:"product_image_url"`
	ProductCurrentPrice string    `json:"product_current_price"`
	ProductCurrency     string    `json:"product_currency"`
	ProductInStock      bool      `json:"product_in_stock"`
	AlertSale           bool      `json:"alert_sale"`
	AlertRestock        bool      `json:"alert_restock"`
	CreatedAt           time.Time `json:"created_at"`
}

func (app *application) addToWatchlistHandler(c *gin.Context) {
	userId := c.MustGet("userId").(uuid.UUID)

	var input struct {
		ProductID    uuid.UUID `json:"product_id" binding:"required"`
		AlertSale    *bool     `json:"alert_sale"`
		AlertRestock *bool     `json:"alert_restock"`
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

	alertSale := true
	if input.AlertSale != nil {
		alertSale = *input.AlertSale
	}
	alertRestock := true
	if input.AlertRestock != nil {
		alertRestock = *input.AlertRestock
	}

	entry, err := app.db.InsertWatchlistEntry(c.Request.Context(), database.InsertWatchlistEntryParams{
		UserID:       userId,
		ProductID:    input.ProductID,
		AlertSale:    alertSale,
		AlertRestock: alertRestock,
	})
	if err != nil {
		var pqErr *pq.Error
		switch {
		case errors.As(err, &pqErr) && pqErr.Code == "23505":
			app.errorResponse(c, http.StatusConflict, "product is already in your watchlist")
		default:
			app.serverErrorResponse(c, err)
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"watchlist_entry": watchlistEntry{
			ID:           entry.ID,
			ProductID:    entry.ProductID,
			AlertSale:    entry.AlertSale,
			AlertRestock: entry.AlertRestock,
			CreatedAt:    entry.CreatedAt.Time,
		},
	})
}

func (app *application) getWatchlistHandler(c *gin.Context) {
	userId := c.MustGet("userId").(uuid.UUID)

	rows, err := app.db.GetWatchlistByUser(c.Request.Context(), userId)
	if err != nil {
		app.serverErrorResponse(c, err)
		return
	}

	result := make([]watchlistEntryWithProduct, len(rows))
	for i, r := range rows {
		result[i] = watchlistEntryWithProduct{
			ID:                  r.ID,
			ProductID:           r.ProductID,
			ProductName:         r.ProductName,
			ProductSlug:         r.ProductSlug,
			ProductImageURL:     r.ProductImageUrl,
			ProductCurrentPrice: r.ProductCurrentPrice,
			ProductCurrency:     r.ProductCurrency,
			ProductInStock:      r.ProductInStock,
			AlertSale:           r.AlertSale,
			AlertRestock:        r.AlertRestock,
			CreatedAt:           r.CreatedAt.Time,
		}
	}

	c.JSON(http.StatusOK, gin.H{"watchlist": result})
}

func (app *application) updateWatchlistEntryHandler(c *gin.Context) {
	userId := c.MustGet("userId").(uuid.UUID)

	entryId, err := uuid.Parse(c.Param("id"))
	if err != nil {
		app.errorResponse(c, http.StatusBadRequest, "invalid watchlist entry id")
		return
	}

	var input struct {
		AlertSale    bool `json:"alert_sale"`
		AlertRestock bool `json:"alert_restock"`
	}

	if err := c.BindJSON(&input); err != nil {
		app.errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	entry, err := app.db.UpdateWatchlistEntry(c.Request.Context(), database.UpdateWatchlistEntryParams{
		AlertSale:    input.AlertSale,
		AlertRestock: input.AlertRestock,
		ID:           entryId,
		UserID:       userId,
	})
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			app.errorResponse(c, http.StatusNotFound, "watchlist entry not found")
		default:
			app.serverErrorResponse(c, err)
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"watchlist_entry": watchlistEntry{
			ID:           entry.ID,
			ProductID:    entry.ProductID,
			AlertSale:    entry.AlertSale,
			AlertRestock: entry.AlertRestock,
			CreatedAt:    entry.CreatedAt.Time,
		},
	})
}

func (app *application) deleteWatchlistEntryHandler(c *gin.Context) {
	userId := c.MustGet("userId").(uuid.UUID)

	entryId, err := uuid.Parse(c.Param("id"))
	if err != nil {
		app.errorResponse(c, http.StatusBadRequest, "invalid watchlist entry id")
		return
	}

	err = app.db.DeleteWatchlistEntry(c.Request.Context(), database.DeleteWatchlistEntryParams{
		ID:     entryId,
		UserID: userId,
	})
	if err != nil {
		app.serverErrorResponse(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}
