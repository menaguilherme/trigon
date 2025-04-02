package main

import (
	"net/http"

	"github.com/menaguilherme/trigon/internal/store"
)

type contextKey string

const (
	userCtxKey contextKey = "user"
	rtvCtxKey  contextKey = "refreshTokenVersion"
)

func getUserFromContext(r *http.Request) *store.User {
	user, _ := r.Context().Value(userCtxKey).(*store.User)
	return user
}

func getRefreshTokenVersionFromContext(r *http.Request) int {
	rtv, _ := r.Context().Value(rtvCtxKey).(int)
	return rtv
}
