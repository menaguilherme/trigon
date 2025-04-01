package main

import (
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	gonanoid "github.com/matoous/go-nanoid"
	"github.com/menaguilherme/trigon/internal/store"
)

type RegisterUserPayload struct {
	FirstName string `json:"first_name" validate:"required,max=80"`
	LastName  string `json:"last_name" validate:"required,max=80"`
	Username  string `json:"username" validate:"required,max=255"`
	Email     string `json:"email" validate:"required,email,max=255"`
	Password  string `json:"password" validate:"required,min=3,max=72"`
}

type AuthInfo struct {
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
	Type         string    `json:"type"`
	ExpiresAt    time.Time `json:"expires_at"`
}

type UserWithAuth struct {
	Auth AuthInfo    `json:"auth"`
	User *store.User `json:"user"`
}

func (app *application) RegisterUserHandler(w http.ResponseWriter, r *http.Request) {
	var payload RegisterUserPayload
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user := &store.User{
		FirstName: payload.FirstName,
		LastName:  payload.LastName,
		Username:  payload.Username,
		Email:     payload.Email,
	}

	if err := user.Password.Set(payload.Password); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	ctx := r.Context()

	err := app.store.Users.Create(ctx, user)
	if err != nil {
		switch err {
		case store.ErrDuplicateEmail:
			app.badRequestResponse(w, r, err)
		case store.ErrDuplicateUsername:
			app.badRequestResponse(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	response := map[string]interface{}{
		"message": "Successfully created user.",
	}

	if err := app.jsonResponse(w, http.StatusCreated, response); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

type LoginPayload struct {
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=3,max=72"`
}

func (app *application) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var payload LoginPayload
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user, err := app.store.Users.GetByEmail(r.Context(), payload.Email)
	if err != nil {
		switch err {
		case store.ErrNotFound:
			app.unauthorizedErrorResponse(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	if err := user.Password.Compare(payload.Password); err != nil {
		app.unauthorizedErrorResponse(w, r, err)
		return
	}

	expiresIn := 15 * time.Minute
	expiresAt := time.Now().Add(expiresIn)
	refreshToken, err := gonanoid.Nanoid(32)
	refreshExpiresAt := time.Now().Add(7 * 24 * time.Hour)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	claims := jwt.MapClaims{
		"sub": user.ID,
		"exp": expiresAt.Unix(),
		"iat": time.Now().Unix(),
		"nbf": time.Now().Unix(),
	}

	accessToken, err := app.authenticator.GenerateToken(claims)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	ctx := r.Context()

	err = app.store.RefreshTokens.Create(ctx, &store.RefreshToken{
		UserID:    user.ID,
		Token:     refreshToken,
		Version:   user.RefreshTokenVersion,
		ExpiresAt: refreshExpiresAt,
	})

	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	response := UserWithAuth{
		Auth: AuthInfo{
			Token:        accessToken,
			RefreshToken: refreshToken,
			Type:         "Bearer",
			ExpiresAt:    refreshExpiresAt,
		},
		User: user,
	}

	if err := app.jsonResponse(w, http.StatusOK, response); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

type RefreshTokenPayload struct {
	RefreshToken string `json:"refresh_token"`
}

func (app *application) RefreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	var payload RefreshTokenPayload
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	tokenRecord, err := app.store.RefreshTokens.GetByToken(r.Context(), payload.RefreshToken)
	if err != nil {
		switch err {
		case store.ErrNotFound:
			app.jsonMessageResponse(w, http.StatusBadRequest, "Invalid refresh token")
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	if time.Now().After(tokenRecord.ExpiresAt) {
		app.jsonMessageResponse(w, http.StatusBadRequest, "Expired refresh token")
		return
	}

	if tokenRecord.RevokedAt.Valid {
		app.jsonMessageResponse(w, http.StatusBadRequest, "Revoked refresh token")
		return
	}

	user, err := app.store.Users.GetByID(r.Context(), tokenRecord.UserID)
	if err != nil {
		switch err {
		case store.ErrNotFound:
			app.jsonResponse(w, http.StatusBadRequest, "User not found")
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	if user.RefreshTokenVersion != tokenRecord.Version {
		app.jsonMessageResponse(w, http.StatusBadRequest, "Invalid refresh token version")
		return
	}

	expiresIn := 15 * time.Minute
	expiresAt := time.Now().Add(expiresIn)
	refreshToken, err := gonanoid.Nanoid(32)
	refreshExpiresAt := time.Now().Add(7 * 24 * time.Hour)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	claims := jwt.MapClaims{
		"sub": user.ID,
		"exp": expiresAt.Unix(),
		"iat": time.Now().Unix(),
		"nbf": time.Now().Unix(),
	}

	accessToken, err := app.authenticator.GenerateToken(claims)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	ctx := r.Context()

	err = app.store.RefreshTokens.RevokeTokenByID(ctx, tokenRecord.ID)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	err = app.store.RefreshTokens.Create(ctx, &store.RefreshToken{
		UserID:    user.ID,
		Token:     refreshToken,
		Version:   user.RefreshTokenVersion,
		ExpiresAt: refreshExpiresAt,
	})

	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	response := UserWithAuth{
		Auth: AuthInfo{
			Token:        accessToken,
			RefreshToken: refreshToken,
			Type:         "Bearer",
			ExpiresAt:    refreshExpiresAt,
		},
		User: user,
	}

	if err := app.jsonResponse(w, http.StatusOK, response); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}
