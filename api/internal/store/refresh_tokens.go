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

func (s *RefreshTokenStore) GetByToken(ctx context.Context, refresh_token string) (*RefreshToken, error) {
	query := `
		SELECT id, user_id, token, version, expires_at, created_at, updated_at, revoked_at
		FROM refresh_tokens
		WHERE token = $1
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	refreshToken := &RefreshToken{}
	err := s.db.QueryRowContext(
		ctx,
		query,
		refresh_token,
	).Scan(
		&refreshToken.ID,
		&refreshToken.UserID,
		&refreshToken.Token,
		&refreshToken.Version,
		&refreshToken.ExpiresAt,
		&refreshToken.CreatedAt,
		&refreshToken.UpdatedAt,
		&refreshToken.RevokedAt,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}

	return refreshToken, nil
}

func (s *RefreshTokenStore) RevokeTokenByID(ctx context.Context, id string) error {
	query := `
		UPDATE refresh_tokens SET revoked_at = $1 WHERE id = $2
	`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := s.db.ExecContext(
		ctx,
		query,
		time.Now(),
		id,
	)

	return err
}
