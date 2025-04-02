package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrDuplicateEmail    = errors.New("a user with that email already exists")
	ErrDuplicateUsername = errors.New("a user with that username already exists")
)

type User struct {
	ID                  string         `json:"id"`
	FirstName           string         `json:"first_name"`
	LastName            string         `json:"last_name"`
	Username            string         `json:"username"`
	Email               string         `json:"email"`
	Password            password       `json:"-"`
	ProfileURL          sql.NullString `json:"profile_url"`
	RefreshTokenVersion int            `json:"refresh_token_version"`
	IsDeleted           bool           `json:"is_deleted"`
	IsBlocked           bool           `json:"is_blocked"`
	DeletedAt           sql.NullString `json:"deleted_at"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
}

type password struct {
	text *string
	hash []byte
}

func (p *password) Set(text string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(text), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	p.text = &text
	p.hash = hash

	return nil
}

func (p *password) Compare(text string) error {
	return bcrypt.CompareHashAndPassword(p.hash, []byte(text))
}

type UserStore struct {
	db *sql.DB
}

func (s *UserStore) Create(ctx context.Context, user *User) error {
	query := `
		INSERT INTO users (id, first_name, last_name, username, email, password, profile_url)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, refresh_token_version, is_deleted, is_blocked, deleted_at, created_at, updated_at
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	userId, err := generateId("user")
	if err != nil {
		return err
	}

	err = s.db.QueryRowContext(
		ctx,
		query,
		userId,
		user.FirstName,
		user.LastName,
		user.Username,
		user.Email,
		user.Password.hash,
		user.ProfileURL,
	).Scan(
		&user.ID,
		&user.RefreshTokenVersion,
		&user.IsDeleted,
		&user.IsBlocked,
		&user.DeletedAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		case err.Error() == `pq: duplicate key value violates unique constraint "users_username_key"`:
			return ErrDuplicateUsername
		default:
			return err
		}
	}

	return nil
}

func (s *UserStore) GetByEmail(ctx context.Context, email string) (*User, error) {
	query := `
		SELECT id, first_name, last_name, username, email, password, profile_url, refresh_token_version, is_deleted, is_blocked, deleted_at, created_at, updated_at
		FROM users
		WHERE email = $1 AND is_blocked = false
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	user := &User{}
	err := s.db.QueryRowContext(
		ctx,
		query,
		email,
	).Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Username,
		&user.Email,
		&user.Password.hash,
		&user.ProfileURL,
		&user.RefreshTokenVersion,
		&user.IsDeleted,
		&user.IsBlocked,
		&user.DeletedAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}

	return user, nil
}

func (s *UserStore) GetByID(ctx context.Context, id string) (*User, error) {
	query := `
		SELECT id, first_name, last_name, username, email, password, profile_url, refresh_token_version, is_deleted, is_blocked, deleted_at, created_at, updated_at
		FROM users
		WHERE id = $1 AND is_blocked = false
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	user := &User{}
	err := s.db.QueryRowContext(
		ctx,
		query,
		id,
	).Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Username,
		&user.Email,
		&user.Password.hash,
		&user.ProfileURL,
		&user.RefreshTokenVersion,
		&user.IsDeleted,
		&user.IsBlocked,
		&user.DeletedAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}

	return user, nil
}

func (s *UserStore) IncreaseTokenVersion(ctx context.Context, user *User) error {
	fmt.Println(user)
	query := `
		UPDATE users 
		SET refresh_token_version = refresh_token_version + 1
		WHERE id = $1
		RETURNING refresh_token_version
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	err := s.db.QueryRowContext(
		ctx,
		query,
		user.ID,
	).Scan(
		&user.RefreshTokenVersion,
	)

	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return ErrNotFound
		default:
			return err
		}
	}

	return nil
}
