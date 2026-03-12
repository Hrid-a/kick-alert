package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/Hrid-a/kick-alert/internal/database"
	"github.com/robfig/cron/v3"
)

func (app *application) startScheduler() {
	c := cron.New()

	c.AddFunc("@every 5m", func() {
		ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
		defer cancel()

		staleProducts, err := app.db.GetStaleProducts(ctx)
		if err != nil {
			app.logger.Error("scheduler: failed to fetch stale products", slog.String("error", err.Error()))
			return
		}

		app.logger.Info("scheduler: scraping stale products", slog.Int("count", len(staleProducts)))

		for _, p := range staleProducts {
			app.wg.Go(func() {
				defer func() {
					pv := recover()
					if pv != nil {
						app.logger.Error(fmt.Sprintf("%v", pv))
					}
				}()

				app.scrapeProduct(ctx, p)
			})
		}
	})

	c.Start()
	app.logger.Info("scheduler: started")
}

func (app *application) scrapeProduct(ctx context.Context, p database.Product) {
	item, err := app.scrapeNikeProduct(ctx, p.Url)
	if err != nil {
		app.logger.Error("scheduler: failed to scrape product", slog.String("product_id", p.ID.String()), slog.String("error", err.Error()))
		return
	}

	currentPrice, err := strconv.ParseFloat(p.CurrentPrice, 64)
	if err != nil {
		app.logger.Error("scheduler: invalid stored price", slog.String("product_id", p.ID.String()), slog.String("value", p.CurrentPrice))
		return
	}

	scrapedPrice := item.Price.Current
	scrapedInStock := item.Status == "BUYABLE_BUY"
	scrapedPriceStr := fmt.Sprintf("%.2f", scrapedPrice)

	priceDropped := scrapedPrice < currentPrice
	priceRise := scrapedPrice > currentPrice
	restocked := scrapedInStock && !p.InStock
	changed := priceDropped || restocked || priceRise

	if !changed {
		// only update last_scraped_at to prevent re-queuing, no history entry needed
		if err := app.db.UpdateProduct(ctx, database.UpdateProductParams{
			ID:           p.ID,
			CurrentPrice: p.CurrentPrice,
			InStock:      p.InStock,
		}); err != nil {
			app.logger.Error("scheduler: failed to update product", slog.String("product_id", p.ID.String()), slog.String("error", err.Error()))
		}
		return
	}

	_, err = app.db.InsertPriceHistory(ctx, database.InsertPriceHistoryParams{
		ProductID: p.ID,
		Price:     scrapedPriceStr,
		InStock:   scrapedInStock,
	})
	if err != nil {
		app.logger.Error("scheduler: failed to insert price history", slog.String("product_id", p.ID.String()), slog.String("error", err.Error()))
		return
	}

	err = app.db.UpdateProduct(ctx, database.UpdateProductParams{
		ID:           p.ID,
		CurrentPrice: scrapedPriceStr,
		InStock:      scrapedInStock,
	})
	if err != nil {
		app.logger.Error("scheduler: failed to update product", slog.String("product_id", p.ID.String()), slog.String("error", err.Error()))
		return
	}

	if priceDropped {
		app.logger.Info("scheduler: price drop detected", slog.String("product_id", p.ID.String()))
		app.notifyWatchers(ctx, p, scrapedPriceStr, database.NotificationTypePRICEDROP)
	}
	if priceRise {
		app.logger.Info("scheduler: price rise detected", slog.String("product_id", p.ID.String()))
	}
	if restocked {
		app.logger.Info("scheduler: restock detected", slog.String("product_id", p.ID.String()))
		app.notifyWatchers(ctx, p, scrapedPriceStr, database.NotificationTypeRESTOCK)
	}
}

func (app *application) notifyWatchers(ctx context.Context, p database.Product, newPrice string, notifType database.NotificationType) {
	watchers, err := app.db.GetWatchersByProduct(ctx, p.ID)
	if err != nil {
		app.logger.Error("notifier: failed to get watchers", slog.String("product_id", p.ID.String()), slog.String("error", err.Error()))
		return
	}

	for _, w := range watchers {
		if notifType == database.NotificationTypePRICEDROP && !w.AlertSale {
			continue
		}
		if notifType == database.NotificationTypeRESTOCK && !w.AlertRestock {
			continue
		}

		_, err := app.db.InsertNotification(ctx, database.InsertNotificationParams{
			UserID:      w.UserID,
			ProductID:   p.ID,
			WatchlistID: w.WatchlistID,
			Type:        notifType,
			OldPrice:    sql.NullString{String: p.CurrentPrice, Valid: true},
			NewPrice:    sql.NullString{String: newPrice, Valid: true},
		})
		if err != nil {
			app.logger.Error("notifier: failed to insert notification", slog.String("product_id", p.ID.String()), slog.String("user_id", w.UserID.String()), slog.String("error", err.Error()))
		}

		// skip email if user has opted out
		if w.NotifyEmail.Valid && !w.NotifyEmail.Bool {
			continue
		}

		err = app.mailer.Send(w.Email, "price_alert.tmpl", map[string]any{
			"name":        w.Name,
			"productName": p.Name,
			"productURL":  p.Url,
			"oldPrice":    p.CurrentPrice,
			"newPrice":    newPrice,
			"currency":    p.Currency,
			"alertType":   string(notifType),
		})
		if err != nil {
			app.logger.Error("notifier: failed to send email", slog.String("product_id", p.ID.String()), slog.String("user_id", w.UserID.String()), slog.String("error", err.Error()))
		}
	}
}
