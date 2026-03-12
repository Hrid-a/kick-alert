package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (app *application) logError(method, url string, err error) {

	app.logger.Error("request error", "error", err.Error(), "method", method, "url", url)
}

func (app *application) errorResponse(c *gin.Context, status int, msg any) {

	c.JSON(status, gin.H{
		"error": msg,
	})
}

func (app *application) serverErrorResponse(c *gin.Context, err error) {
	app.logError(c.Request.Method, c.Request.RequestURI, err)
	message := "The server encounterd a problem and could not process your request"
	app.errorResponse(c, http.StatusInternalServerError, message)
}

func (app *application) rateLimitExceededResponse(c *gin.Context) {
	message := "rate limit exceeded"
	app.errorResponse(c, http.StatusTooManyRequests, message)
}
