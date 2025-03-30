package store

import (
	"context"
	"database/sql"
)

type User struct {
	ID string `json:"id"`
}

type UserStore struct {
	db *sql.DB
}

func (s *UserStore) Create(ctx context.Context) error {
	return nil
}
