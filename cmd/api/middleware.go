package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/Hrid-a/kick-alert/internal/auth"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/time/rate"
)

func (app *application) authenticate() gin.HandlerFunc {

	return func(c *gin.Context) {

		tokenStr, err := auth.GetBearerToken(c.Request.Header)

		if err != nil {
			app.errorResponse(c, http.StatusUnauthorized, err.Error())
			c.Abort()
			return
		}

		userId, err := auth.ValidateJWT(tokenStr, app.config.jwt_secret)

		if err != nil || userId == uuid.Nil {
			fmt.Printf("\n\n invalid JWT\n\n")
			app.errorResponse(c, http.StatusUnauthorized, "invalid or missing authentication token")
			c.Abort()
			return
		}

		c.Set("userId", userId)
		c.Next()
	}
}

func (app *application) rateLimit() gin.HandlerFunc {

	if !app.config.limiter.enabled {
		return func(c *gin.Context) { c.Next() }
	}

	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}

	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	)

	go func() {
		for {
			time.Sleep(time.Minute)

			mu.Lock()

			for ip, client := range clients {
				if time.Since(client.lastSeen) > 3*time.Minute {
					delete(clients, ip)
				}
			}

			mu.Unlock()
		}
	}()

	return func(c *gin.Context) {
		ip := c.ClientIP()

		mu.Lock()

		if _, found := clients[ip]; !found {
			clients[ip] = &client{
				// Use the requests-per-second and burst values from the config
				// struct.
				limiter: rate.NewLimiter(rate.Limit(app.config.limiter.rps), app.config.limiter.burst),
			}
		}

		clients[ip].lastSeen = time.Now()

		if !clients[ip].limiter.Allow() {
			mu.Unlock()
			app.rateLimitExceededResponse(c)
			return
		}

		mu.Unlock()

		c.Next()
	}

}
