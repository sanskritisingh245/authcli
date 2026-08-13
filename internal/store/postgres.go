package store

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"authcli/internal/models"
)

var ErrNotFound = errors.New("not found")

type Store struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

func (s *Store) CreateUser(ctx context.Context, username, passwordHash string) (models.User, error) {
	u := models.User{Username: username, PasswordHash: passwordHash}
	err := s.pool.QueryRow(ctx,
		`INSERT INTO users (username, password_hash) VALUES ($1, $2)
		 RETURNING id, created_at`,
		username, passwordHash,
	).Scan(&u.ID, &u.CreatedAt)
	return u, err
}

func (s *Store) FindByUsername(ctx context.Context, username string) (models.User, error) {
	var u models.User
	err := s.pool.QueryRow(ctx,
		`SELECT id, username, password_hash, failed_attempts, locked_until, last_login_at, created_at
		 FROM users WHERE username = $1`,
		username,
	).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.FailedAttempts, &u.LockedUntil, &u.LastLoginAt, &u.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.User{}, ErrNotFound
	}
	return u, err
}

func (s *Store) GetUserByID(ctx context.Context, id int64) (models.User, error) {
	var u models.User
	err := s.pool.QueryRow(ctx,
		`SELECT id, username, password_hash, failed_attempts, locked_until, last_login_at, created_at
		 FROM users WHERE id = $1`,
		id,
	).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.FailedAttempts, &u.LockedUntil, &u.LastLoginAt, &u.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.User{}, ErrNotFound
	}
	return u, err
}

func (s *Store) IncrementFailedAttempts(ctx context.Context, userID int64) (int, error) {
	var attempts int
	err := s.pool.QueryRow(ctx,
		`UPDATE users SET failed_attempts = failed_attempts + 1 WHERE id = $1 RETURNING failed_attempts`,
		userID,
	).Scan(&attempts)
	return attempts, err
}

func (s *Store) SetLockout(ctx context.Context, userID int64, until time.Time) error {
	_, err := s.pool.Exec(ctx, `UPDATE users SET locked_until = $1 WHERE id = $2`, until, userID)
	return err
}

func (s *Store) ResetLoginState(ctx context.Context, userID int64, loginAt time.Time) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE users SET failed_attempts = 0, locked_until = NULL, last_login_at = $1 WHERE id = $2`,
		loginAt, userID,
	)
	return err
}

func (s *Store) CreateSession(ctx context.Context, userID int64, tokenHash string, expiresAt time.Time) error {
	_, err := s.pool.Exec(ctx,
		`INSERT INTO sessions (user_id, token_hash, expires_at) VALUES ($1, $2, $3)`,
		userID, tokenHash, expiresAt,
	)
	return err
}

func (s *Store) FindSessionByTokenHash(ctx context.Context, tokenHash string) (models.Session, error) {
	var sess models.Session
	err := s.pool.QueryRow(ctx,
		`SELECT id, user_id, token_hash, created_at, expires_at, revoked_at
		 FROM sessions
		 WHERE token_hash = $1 AND revoked_at IS NULL AND expires_at > now()`,
		tokenHash,
	).Scan(&sess.ID, &sess.UserID, &sess.TokenHash, &sess.CreatedAt, &sess.ExpiresAt, &sess.RevokedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Session{}, ErrNotFound
	}
	return sess, err
}

func (s *Store) RevokeSession(ctx context.Context, tokenHash string) error {
	_, err := s.pool.Exec(ctx, `UPDATE sessions SET revoked_at = now() WHERE token_hash = $1`, tokenHash)
	return err
}

func (s *Store) SaveTOTPSecret(ctx context.Context, userID int64, secret string) error {
	_, err := s.pool.Exec(ctx,
		`INSERT INTO totp_secrets (user_id, secret, enabled) VALUES ($1, $2, false)
		 ON CONFLICT (user_id) DO UPDATE SET secret = $2, enabled = false`,
		userID, secret,
	)
	return err
}

func (s *Store) EnableTOTP(ctx context.Context, userID int64) error {
	_, err := s.pool.Exec(ctx, `UPDATE totp_secrets SET enabled = true WHERE user_id = $1`, userID)
	return err
}

func (s *Store) FindTOTPSecret(ctx context.Context, userID int64) (models.TOTPSecret, error) {
	var t models.TOTPSecret
	err := s.pool.QueryRow(ctx,
		`SELECT user_id, secret, enabled, created_at FROM totp_secrets WHERE user_id = $1`,
		userID,
	).Scan(&t.UserID, &t.Secret, &t.Enabled, &t.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.TOTPSecret{}, ErrNotFound
	}
	return t, err
}

func (s *Store) DeleteTOTPSecret(ctx context.Context, userID int64) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM totp_secrets WHERE user_id = $1`, userID)
	return err
}
