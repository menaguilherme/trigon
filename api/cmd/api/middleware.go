package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/menaguilherme/trigon/internal/store"
)

func (app *application) AuthTokenMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			app.unauthorizedErrorResponse(w, r, fmt.Errorf("authorization header is missing"))
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			app.unauthorizedErrorResponse(w, r, fmt.Errorf("authorization header is malformed"))
			return
		}

		token := parts[1]
		jwtToken, err := app.authenticator.ValidateToken(token)
		if err != nil {
			app.unauthorizedErrorResponse(w, r, err)
			return
		}

		claims, _ := jwtToken.Claims.(jwt.MapClaims)

		userID, err := claims.GetSubject()
		if err != nil {
			app.unauthorizedErrorResponse(w, r, err)
			return
		}

		rtvFloat, ok := claims["rtv"].(float64)
		if !ok {
			app.unauthorizedErrorResponse(w, r, fmt.Errorf("refresh token version claim is missing or invalid"))
			return
		}
		rtv := int(rtvFloat)

		user, err := app.getUser(r.Context(), userID)
		if err != nil {
			app.unauthorizedErrorResponse(w, r, err)
			return
		}

		if rtv != user.RefreshTokenVersion {
			app.unauthorizedErrorResponse(w, r, fmt.Errorf("invalid token: version mismatch"))
			return
		}

		ctx := context.WithValue(r.Context(), userCtxKey, user)
		ctx = context.WithValue(ctx, rtvCtxKey, rtv)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (app *application) getUser(ctx context.Context, userID string) (*store.User, error) {

	user, err := app.store.Users.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return user, nil
}
