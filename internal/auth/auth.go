package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type TokenScope string

const (
	TokenTypeAccess     TokenScope = "kick-alert-access"
	TokenTypeRefresh    TokenScope = "kick-alert-refresh"
	TokenTypeActivation TokenScope = "kick-alert-activate"
)

var ErrNoAuthHeaderIncluded = errors.New("no auth header included in request")

func HashPassword(password string) (string, error) {
	hash, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return "", err
	}

	return hash, nil
}

func CheckPassword(password, hash string) (bool, error) {
	match, err := argon2id.ComparePasswordAndHash(password, hash)

	if err != nil {
		return false, err
	}

	return match, nil
}

func MakeJWT(id uuid.UUID, expiresIn time.Duration, tokenSecret string) (string, error) {

	key := []byte(tokenSecret)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    string(TokenTypeAccess),
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
		Subject:   id.String(),
	})

	return token.SignedString(key)
}

func ValidateJWT(tokenStr, tokenSecret string) (uuid.UUID, error) {
	claims := jwt.RegisteredClaims{}

	token, err := jwt.ParseWithClaims(tokenStr, &claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(tokenSecret), nil
	})

	if err != nil || !token.Valid {
		return uuid.Nil, fmt.Errorf("invalid or missing authentication token")
	}

	userIdStr, err := token.Claims.GetSubject()

	if err != nil {
		return uuid.Nil, err
	}

	issuer, err := token.Claims.GetIssuer()

	if err != nil {
		return uuid.Nil, err
	}

	if issuer != string(TokenTypeAccess) {
		return uuid.Nil, fmt.Errorf("invalid issuer")
	}

	id, err := uuid.Parse(userIdStr)

	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid userId %w", err)
	}

	return id, nil
}

func GetBearerToken(header http.Header) (string, error) {

	authHeader := header.Get("Authorization")

	if authHeader == "" {
		return "", ErrNoAuthHeaderIncluded
	}

	splitAuth := strings.Split(authHeader, " ")

	if len(splitAuth) < 2 || splitAuth[0] != "Bearer" {
		return "", errors.New("invalid or missing authentication token")
	}

	return splitAuth[1], nil
}

func MakeRefreshToken() (string, error) {
	token := make([]byte, 32)
	if _, err := rand.Read(token); err != nil {
		return "", err
	}
	return hex.EncodeToString(token), nil
}

func GenerateToken(plainText string) string {

	hash := sha256.Sum256([]byte(plainText))

	return hex.EncodeToString(hash[:])
}
