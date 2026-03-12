package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (app *application) healthceckHandler(c *gin.Context) {

	c.JSON(http.StatusOK, gin.H{
		"status": "available",
		"system_info": map[string]string{
			"environment": app.config.env,
			"version":     version,
		},
	})
}
