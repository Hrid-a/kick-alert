package main

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/Hrid-a/kick-alert/internal/database"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type notificationResponse struct {
	ID              uuid.UUID                 `json:"id"`
	ProductID       uuid.UUID                 `json:"product_id"`
	ProductName     string                    `json:"product_name"`
	ProductSlug     string                    `json:"product_slug"`
	ProductImageURL string                    `json:"product_image_url"`
	Type            database.NotificationType `json:"type"`
	OldPrice        *string                   `json:"old_price,omitempty"`
	NewPrice        *string                   `json:"new_price,omitempty"`
	Read            bool                      `json:"read"`
	CreatedAt       time.Time                 `json:"created_at"`
}

func (app *application) getNotificationsHandler(c *gin.Context) {
	userId := c.MustGet("userId").(uuid.UUID)

	// ?unread=true  → only unread; omitted → all
	var readFilter sql.NullBool
	if raw := c.Query("unread"); raw != "" {
		unread, err := strconv.ParseBool(raw)
		if err != nil {
			app.errorResponse(c, http.StatusBadRequest, "invalid value for 'unread' query param")
			return
		}
		readFilter = sql.NullBool{Bool: !unread, Valid: true}
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	rows, err := app.db.GetNotificationsByUser(c.Request.Context(), database.GetNotificationsByUserParams{
		UserID:     userId,
		ReadFilter: readFilter,
		Limit:      int32(pageSize),
		Offset:     int32(offset),
	})
	if err != nil {
		app.serverErrorResponse(c, err)
		return
	}

	total, err := app.db.CountNotificationsByUser(c.Request.Context(), database.CountNotificationsByUserParams{
		UserID:     userId,
		ReadFilter: readFilter,
	})
	if err != nil {
		app.serverErrorResponse(c, err)
		return
	}

	result := make([]notificationResponse, len(rows))
	for i, r := range rows {
		n := notificationResponse{
			ID:              r.ID,
			ProductID:       r.ProductID,
			ProductName:     r.ProductName,
			ProductSlug:     r.ProductSlug,
			ProductImageURL: r.ProductImageUrl,
			Type:            r.Type,
			Read:            r.Read,
			CreatedAt:       r.CreatedAt,
		}
		if r.OldPrice.Valid {
			n.OldPrice = &r.OldPrice.String
		}
		if r.NewPrice.Valid {
			n.NewPrice = &r.NewPrice.String
		}
		result[i] = n
	}

	c.JSON(http.StatusOK, gin.H{
		"notifications": result,
		"meta": gin.H{
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

func (app *application) markNotificationReadHandler(c *gin.Context) {
	userId := c.MustGet("userId").(uuid.UUID)

	notifId, err := uuid.Parse(c.Param("id"))
	if err != nil {
		app.errorResponse(c, http.StatusBadRequest, "invalid notification id")
		return
	}

	notif, err := app.db.MarkNotificationRead(c.Request.Context(), database.MarkNotificationReadParams{
		ID:     notifId,
		UserID: userId,
	})
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			app.errorResponse(c, http.StatusNotFound, "notification not found")
		default:
			app.serverErrorResponse(c, err)
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"notification": gin.H{"id": notif.ID, "read": notif.Read}})
}

func (app *application) markAllNotificationsReadHandler(c *gin.Context) {
	userId := c.MustGet("userId").(uuid.UUID)

	if err := app.db.MarkAllNotificationsRead(c.Request.Context(), userId); err != nil {
		app.serverErrorResponse(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}
