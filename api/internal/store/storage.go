package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	gonanoid "github.com/matoous/go-nanoid"
)

var (
	ErrNotFound          = errors.New("resource not found")
	ErrConflict          = errors.New("resource already exists")
	QueryTimeoutDuration = time.Second * 5
)

type Storage struct {
	Users interface {
		Create(context.Context, *User) error
		GetByEmail(ctx context.Context, email string) (*User, error)
		GetByID(context.Context, string) (*User, error)
	}
	RefreshTokens interface {
		Create(context.Context, *RefreshToken) error
		GetByToken(context.Context, string) (*RefreshToken, error)
		RevokeTokenByID(context.Context, string) error
	}
}

func NewStorage(db *sql.DB) Storage {
	return Storage{
		Users:         &UserStore{db},
		RefreshTokens: &RefreshTokenStore{db},
	}
}

func withTx(db *sql.DB, ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}

func generateId(prefix string) (string, error) {
	customAlphabet := "abcdefghijklmnopqrstuvwxyz0123456789"
	id, err := gonanoid.Generate(customAlphabet, 22)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s_%s", prefix, id), nil

}
