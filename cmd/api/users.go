package main

import (
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/Hrid-a/kick-alert/internal/auth"
	"github.com/Hrid-a/kick-alert/internal/database"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

type User struct {
	ID        uuid.UUID    `json:"id"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
	Email     string       `json:"email"`
	Name      string       `json:"name"`
	Password  string       `json:"-"`
	Activated sql.NullBool `json:"is_active"`
}

func (app *application) registerUserHandler(c *gin.Context) {

	var input struct {
		Name     string `json:"name" binding:"required,min=2,max=50"`
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=8"`
	}

	if err := c.BindJSON(&input); err != nil {
		app.errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	passwordHash, err := auth.HashPassword(input.Password)

	if err != nil {
		app.serverErrorResponse(c, err)
		return
	}

	user, err := app.db.InsertUser(c.Request.Context(), database.InsertUserParams{
		Email:        input.Email,
		Name:         input.Name,
		PasswordHash: passwordHash,
	})

	if err != nil {
		var pqErr *pq.Error
		switch {
		case errors.As(err, &pqErr) && pqErr.Code == "23505":
			app.errorResponse(c, http.StatusBadRequest, "a user with this email address already exists")
			return
		default:
			app.serverErrorResponse(c, err)
		}
		return
	}

	plainText := rand.Text()
	token := auth.GenerateToken(plainText)
	err = app.db.InsertToken(c.Request.Context(), database.InsertTokenParams{
		Hash:   token,
		UserID: user.ID,
		Expiry: time.Now().Add(30 * time.Minute),
		Scope:  string(auth.TokenTypeActivation),
	})

	if err != nil {
		app.serverErrorResponse(c, err)
		return
	}

	app.wg.Go(func() {

		defer func() {
			pv := recover()
			if pv != nil {
				app.logger.Error(fmt.Sprintf("%v", pv))
			}
		}()

		data := map[string]any{
			"activationToken": plainText,
			"userID":          user.ID,
			"name":            user.Name,
			"redirectURL":     fmt.Sprintf("%s?token=%s", app.config.front_end_activationURL, plainText),
		}

		// Send the welcome email, passing in the map above as dynamic data.
		err := app.mailer.Send(user.Email, "user_welcome.tmpl", data)
		if err != nil {
			app.logger.Error(err.Error())
		}
	})

	c.JSON(http.StatusAccepted, gin.H{
		"user": User{
			ID:        user.ID,
			CreatedAt: user.CreatedAt.Time.UTC(),
			UpdatedAt: user.UpdatedAt.Time,
			Name:      user.Name,
			Email:     user.Email,
		},
	})
}
