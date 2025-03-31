package store

import (
	"context"
	"database/sql"
	"time"
)

type RefreshToken struct {
	ID        string         `json:"id"`
	UserID    string         `json:"user_id"`
	Token     string         `json:"token"`
	Version   int            `json:"version"`
	ExpiresAt time.Time      `json:"expires_at"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	RevokedAt sql.NullString `json:"revoked_at"`
}

type RefreshTokenStore struct {
	db *sql.DB
}

func (s *RefreshTokenStore) Create(ctx context.Context, refresh_token *RefreshToken) error {
	query := `
	INSERT INTO refresh_tokens (id, user_id, token, version, expires_at)
	VALUES ($1, $2, $3, $4, $5)
	RETURNING created_at, updated_at, revoked_at
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	refreshTokenId, err := generateId("reftoken")
	if err != nil {
		return err
	}

	err = s.db.QueryRowContext(
		ctx,
		query,
		refreshTokenId,
		refresh_token.UserID,
		refresh_token.Token,
		refresh_token.Version,
		refresh_token.ExpiresAt,
	).Scan(
		&refresh_token.CreatedAt,
		&refresh_token.UpdatedAt,
		&refresh_token.RevokedAt,
	)

	refresh_token.ID = refreshTokenId

	if err != nil {
		return err
	}

	return nil

}
