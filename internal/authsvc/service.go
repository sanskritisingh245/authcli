package authsvc

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/pquerna/otp"

	"authcli/internal/config"
	"authcli/internal/models"
	"authcli/internal/store"
)

var (
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrAccountLocked      = errors.New("account is locked, try again later")
	ErrUsernameTaken      = errors.New("username already taken")
	ErrPasswordTooShort   = errors.New("password must be at least 8 characters")
	ErrInvalidTOTP        = errors.New("invalid 2fa code")
	ErrTOTPNotSetUp       = errors.New("2fa has not been set up")
	ErrTOTPNotEnabled     = errors.New("2fa is not enabled")
	ErrTOTPAlreadyEnabled = errors.New("2fa is already enabled")
	ErrSessionInvalid     = errors.New("session is invalid or expired")
)

type TOTPPrompt func() (string, error)

type Service struct {
	store *store.Store
	cfg   config.Config
}

func New(st *store.Store, cfg config.Config) *Service {
	return &Service{store: st, cfg: cfg}
}

func normalizeUsername(u string) string {
	return strings.ToLower(strings.TrimSpace(u))
}

func (s *Service) Register(ctx context.Context, username, password string) error {
	username = normalizeUsername(username)
	if len(password) < 8 {
		return ErrPasswordTooShort
	}

	_, err := s.store.FindByUsername(ctx, username)
	if err == nil {
		return ErrUsernameTaken
	}
	if !errors.Is(err, store.ErrNotFound) {
		return err
	}

	hash, err := hashPassword(password)
	if err != nil {
		return err
	}
	_, err = s.store.CreateUser(ctx, username, hash)
	return err
}

func (s *Service) Login(ctx context.Context, username, password string, promptTOTP TOTPPrompt) (models.User, string, error) {
	username = normalizeUsername(username)

	user, err := s.store.FindByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return models.User{}, "", ErrInvalidCredentials
		}
		return models.User{}, "", err
	}

	if isLocked(user) {
		return models.User{}, "", ErrAccountLocked
	}

	if !checkPassword(user.PasswordHash, password) {
		if err := s.registerFailedAttempt(ctx, user.ID); err != nil {
			return models.User{}, "", err
		}
		return models.User{}, "", ErrInvalidCredentials
	}

	secret, err := s.store.FindTOTPSecret(ctx, user.ID)
	switch {
	case err == nil && secret.Enabled:
		code, err := promptTOTP()
		if err != nil {
			return models.User{}, "", err
		}
		if !validateTOTPCode(secret.Secret, code) {
			if err := s.registerFailedAttempt(ctx, user.ID); err != nil {
				return models.User{}, "", err
			}
			return models.User{}, "", ErrInvalidTOTP
		}
	case err == nil, errors.Is(err, store.ErrNotFound):
		// no active 2fa, password alone is sufficient
	default:
		return models.User{}, "", err
	}

	now := time.Now()
	if err := s.store.ResetLoginState(ctx, user.ID, now); err != nil {
		return models.User{}, "", err
	}
	user.FailedAttempts = 0
	user.LockedUntil = nil
	user.LastLoginAt = &now

	token, err := generateToken()
	if err != nil {
		return models.User{}, "", err
	}
	expiresAt := now.Add(s.cfg.SessionDuration())
	if err := s.store.CreateSession(ctx, user.ID, hashToken(token), expiresAt); err != nil {
		return models.User{}, "", err
	}

	return user, token, nil
}

func (s *Service) Logout(ctx context.Context, token string) error {
	return s.store.RevokeSession(ctx, hashToken(token))
}

func (s *Service) ValidateSession(ctx context.Context, token string) (models.User, error) {
	session, err := s.store.FindSessionByTokenHash(ctx, hashToken(token))
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return models.User{}, ErrSessionInvalid
		}
		return models.User{}, err
	}
	return s.store.GetUserByID(ctx, session.UserID)
}

func (s *Service) EnableTOTP(ctx context.Context, userID int64, username string) (*otp.Key, error) {
	existing, err := s.store.FindTOTPSecret(ctx, userID)
	if err == nil && existing.Enabled {
		return nil, ErrTOTPAlreadyEnabled
	}
	if err != nil && !errors.Is(err, store.ErrNotFound) {
		return nil, err
	}

	key, err := generateTOTPSecret(s.cfg.TOTPIssuer, username)
	if err != nil {
		return nil, err
	}
	if err := s.store.SaveTOTPSecret(ctx, userID, key.Secret()); err != nil {
		return nil, err
	}
	return key, nil
}

func (s *Service) ConfirmTOTP(ctx context.Context, userID int64, code string) error {
	secret, err := s.store.FindTOTPSecret(ctx, userID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return ErrTOTPNotSetUp
		}
		return err
	}
	if secret.Enabled {
		return ErrTOTPAlreadyEnabled
	}
	if !validateTOTPCode(secret.Secret, code) {
		return ErrInvalidTOTP
	}
	return s.store.EnableTOTP(ctx, userID)
}

func (s *Service) DisableTOTP(ctx context.Context, userID int64, code string) error {
	secret, err := s.store.FindTOTPSecret(ctx, userID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return ErrTOTPNotEnabled
		}
		return err
	}
	if !secret.Enabled {
		return ErrTOTPNotEnabled
	}
	if !validateTOTPCode(secret.Secret, code) {
		return ErrInvalidTOTP
	}
	return s.store.DeleteTOTPSecret(ctx, userID)
}
