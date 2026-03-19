package main

import (
	"database/sql"
	"errors"
	"net/http"
	"time"

	"github.com/Hrid-a/kick-alert/internal/auth"
	"github.com/Hrid-a/kick-alert/internal/database"
	"github.com/gin-gonic/gin"
)

func (app *application) setRefreshCookie(c *gin.Context, value string, maxAge int) {
	production := app.config.env == "production"
	sameSite := http.SameSiteLaxMode
	if production {
		sameSite = http.SameSiteNoneMode
	}
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "refresh_token",
		Value:    value,
		MaxAge:   maxAge,
		Path:     "/",
		HttpOnly: true,
		Secure:   production,
		SameSite: sameSite,
	})
}

func (app *application) loginUserHandler(c *gin.Context) {

	var input struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=8,max=16"`
	}

	type response struct {
		User  User   `json:"user"`
		Token string `json:"token"`
	}

	if err := c.BindJSON(&input); err != nil {
		app.errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	user, err := app.db.GetUserByEmail(c.Request.Context(), input.Email)

	if err != nil {
		app.errorResponse(c, http.StatusUnauthorized, "incorrect email or password")
		return
	}

	match, err := auth.CheckPassword(input.Password, user.PasswordHash)

	if err != nil || !match {
		app.errorResponse(c, http.StatusUnauthorized, "incorrect email or password")
		return
	}

	accessToken, err := auth.MakeJWT(user.ID, time.Hour, app.config.jwt_secret)

	if err != nil {
		app.serverErrorResponse(c, err)
		return
	}

	refreshToken, err := auth.MakeRefreshToken()

	if err != nil {
		app.serverErrorResponse(c, err)
		return
	}

	err = app.db.InsertToken(c.Request.Context(), database.InsertTokenParams{
		Hash:   refreshToken,
		UserID: user.ID,
		Expiry: time.Now().Add(3 * 24 * time.Hour),
		Scope:  string(auth.TokenTypeRefresh),
	})

	if err != nil {
		app.serverErrorResponse(c, err)
		return
	}

	app.setRefreshCookie(c, refreshToken, 3*24*60*60)

	c.JSON(http.StatusOK, response{
		User: User{
			ID:        user.ID,
			Name:      user.Name,
			Email:     user.Email,
			CreatedAt: user.CreatedAt.Time,
		},
		Token: accessToken,
	})

}

func (app *application) refreshTokenHandler(c *gin.Context) {

	token, err := c.Cookie("refresh_token")
	if err != nil {
		app.errorResponse(c, http.StatusUnauthorized, "missing refresh token")
		return
	}

	user, err := app.db.GetUserToken(c.Request.Context(), database.GetUserTokenParams{
		Hash:   token,
		Scope:  string(auth.TokenTypeRefresh),
		Expiry: time.Now(),
	})

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			app.errorResponse(c, http.StatusBadRequest, "invalid or expired refresh token")
		default:
			app.serverErrorResponse(c, err)
		}
		return
	}

	err = app.db.DeleteAllForUser(c.Request.Context(), database.DeleteAllForUserParams{
		Scope:  string(auth.TokenTypeRefresh),
		UserID: user.ID,
	})

	if err != nil {
		app.serverErrorResponse(c, err)
		return
	}

	accessToken, err := auth.MakeJWT(user.ID, time.Hour, app.config.jwt_secret)

	if err != nil {
		app.serverErrorResponse(c, err)
		return
	}

	refreshToken, err := auth.MakeRefreshToken()

	if err != nil {
		app.serverErrorResponse(c, err)
		return
	}

	err = app.db.InsertToken(c.Request.Context(), database.InsertTokenParams{
		Hash:   refreshToken,
		UserID: user.ID,
		Expiry: time.Now().Add(3 * 24 * time.Hour),
		Scope:  string(auth.TokenTypeRefresh),
	})

	if err != nil {
		app.serverErrorResponse(c, err)
		return
	}

	app.setRefreshCookie(c, refreshToken, 3*24*60*60)

	c.JSON(http.StatusOK, gin.H{
		"access_token": accessToken,
	})
}

func (app *application) logoutHandler(c *gin.Context) {
	app.setRefreshCookie(c, "", -1)
	c.Status(http.StatusNoContent)
}

func (app *application) activateUserHandler(c *gin.Context) {
	// Parse the plaintext activation token from the request body.
	var input struct {
		TokenPlaintext string `json:"token"`
	}

	err := c.BindJSON(&input)
	if err != nil {
		app.errorResponse(c, http.StatusBadRequest, "bad Request")
		return
	}

	token := auth.GenerateToken(input.TokenPlaintext)
	user, err := app.db.GetUserToken(c.Request.Context(), database.GetUserTokenParams{
		Hash:   token,
		Scope:  string(auth.TokenTypeActivation),
		Expiry: time.Now(),
	})

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			app.errorResponse(c, http.StatusBadRequest, "invalid or expired activation token")
		default:
			app.serverErrorResponse(c, err)
		}
		return
	}

	// Save the updated user record in our database, checking for any edit conflicts in
	// the same way that we did for our movie records.
	err = app.db.UpdateUser(c.Request.Context(), database.UpdateUserParams{
		ID:           user.ID,
		Name:         user.Name,
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
		Activated:    sql.NullBool{Bool: true, Valid: true},
	})

	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			app.errorResponse(c, http.StatusBadRequest, "duplicate email")
		case errors.Is(err, sql.ErrNoRows):
			app.errorResponse(c, http.StatusNotFound, "not found")
		default:
			app.serverErrorResponse(c, err)
		}
		return
	}

	err = app.db.DeleteAllForUser(c.Request.Context(), database.DeleteAllForUserParams{
		Scope:  string(auth.TokenTypeActivation),
		UserID: user.ID,
	})

	if err != nil {
		app.serverErrorResponse(c, err)
		return
	}

	// Send the updated user details to the client in a JSON response.
	c.JSON(http.StatusOK, gin.H{"user": User{
		ID:        user.ID,
		CreatedAt: user.CreatedAt.Time.UTC(),
		Name:      user.Name,
		Email:     user.Email,
		Activated: true,
	}})
}
