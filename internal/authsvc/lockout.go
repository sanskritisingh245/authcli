package authsvc

import (
	"context"
	"time"

	"authcli/internal/models"
)

func isLocked(u models.User) bool {
	return u.LockedUntil != nil && u.LockedUntil.After(time.Now())
}

func (s *Service) registerFailedAttempt(ctx context.Context, userID int64) error {
	attempts, err := s.store.IncrementFailedAttempts(ctx, userID)
	if err != nil {
		return err
	}
	if attempts >= s.cfg.LockoutMaxAttempts {
		return s.store.SetLockout(ctx, userID, time.Now().Add(s.cfg.LockoutDuration()))
	}
	return nil
}
