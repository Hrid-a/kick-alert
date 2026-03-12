package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (app *application) routes() *gin.Engine {

	router := gin.Default()

	router.Use(func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 1<<20) // 1 MB
		c.Next()
	})
	router.Use(app.rateLimit())
	router.GET("/v1/healthcheck", app.healthceckHandler)

	// auth
	router.POST("/v1/register", app.registerUserHandler)
	router.POST("/v1/login", app.loginUserHandler)
	router.GET("/v1/refresh", app.refreshTokenHandler)
	router.PUT("/v1/activation", app.activateUserHandler)

	// Products
	router.POST("/v1/products", app.authenticate(), app.addProductHandler)
	router.GET("/v1/products", app.authenticate(), app.getProductsHandler)
	router.GET("/v1/products/:id", app.authenticate(), app.getProductHandler)

	// Watchlist
	router.POST("/v1/watchlist", app.authenticate(), app.addToWatchlistHandler)
	router.GET("/v1/watchlist", app.authenticate(), app.getWatchlistHandler)
	router.PATCH("/v1/watchlist/:id", app.authenticate(), app.updateWatchlistEntryHandler)
	router.DELETE("/v1/watchlist/:id", app.authenticate(), app.deleteWatchlistEntryHandler)

	// Notifications
	router.GET("/v1/notifications", app.authenticate(), app.getNotificationsHandler)
	router.PATCH("/v1/notifications/read-all", app.authenticate(), app.markAllNotificationsReadHandler)
	router.PATCH("/v1/notifications/:id/read", app.authenticate(), app.markNotificationReadHandler)
	return router
}
