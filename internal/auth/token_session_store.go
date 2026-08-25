package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"law-oa-go/internal/models"
)

var ErrTokenSessionStoreUnavailable = errors.New("auth token session store unavailable")

// TokenSessionStore owns durable JWT session state in PostgreSQL.
type TokenSessionStore struct {
	db *gorm.DB
}

func NewTokenSessionStore(db *gorm.DB) *TokenSessionStore {
	return &TokenSessionStore{db: db}
}

func (s *TokenSessionStore) Create(ctx context.Context, session *models.AuthTokenSession) error {
	if s == nil || s.db == nil {
		return ErrTokenSessionStoreUnavailable
	}
	if err := s.db.WithContext(ctx).Create(session).Error; err != nil {
		return fmt.Errorf("create auth token session: %w", err)
	}
	return nil
}

func (s *TokenSessionStore) GetByAccessUUID(ctx context.Context, uuid string) (*models.AuthTokenSession, error) {
	return s.get(ctx, "access_token_uuid = ?", uuid)
}

func (s *TokenSessionStore) GetByRefreshUUID(ctx context.Context, uuid string) (*models.AuthTokenSession, error) {
	return s.get(ctx, "refresh_token_uuid = ?", uuid)
}

func (s *TokenSessionStore) get(ctx context.Context, query string, value string) (*models.AuthTokenSession, error) {
	if s == nil || s.db == nil {
		return nil, ErrTokenSessionStoreUnavailable
	}
	var session models.AuthTokenSession
	if err := s.db.WithContext(ctx).Where(query, value).First(&session).Error; err != nil {
		return nil, err
	}
	return &session, nil
}

func (s *TokenSessionStore) RevokeTokenUUID(ctx context.Context, uuid, tokenType string, revokedAt time.Time) (int64, error) {
	if s == nil || s.db == nil {
		return 0, ErrTokenSessionStoreUnavailable
	}

	query := s.db.WithContext(ctx).Model(&models.AuthTokenSession{}).Where("revoked_at IS NULL")
	switch tokenType {
	case "access":
		query = query.Where("access_token_uuid = ? AND access_revoked_at IS NULL", uuid)
	case "refresh":
		query = query.Where("refresh_token_uuid = ? AND refresh_revoked_at IS NULL", uuid)
	default:
		return 0, fmt.Errorf("unknown token type %q", tokenType)
	}

	updates := map[string]interface{}{"revoked_at": revokedAt}
	if tokenType == "access" {
		updates["access_revoked_at"] = revokedAt
	} else {
		updates["refresh_revoked_at"] = revokedAt
	}
	result := query.Updates(updates)
	if result.Error != nil {
		return 0, fmt.Errorf("revoke auth token: %w", result.Error)
	}
	return result.RowsAffected, nil
}

func (s *TokenSessionStore) RevokeAllForUser(ctx context.Context, userID uint, revokedAt time.Time) error {
	if s == nil || s.db == nil {
		return ErrTokenSessionStoreUnavailable
	}
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := lockUserForTokenMutation(tx, userID); err != nil {
			return err
		}
		return tx.Model(&models.AuthTokenSession{}).
			Where("user_id = ? AND revoked_at IS NULL", userID).
			Updates(map[string]interface{}{
				"access_revoked_at":  revokedAt,
				"refresh_revoked_at": revokedAt,
				"revoked_at":         revokedAt,
			}).Error
	})
	if err != nil {
		return fmt.Errorf("revoke all auth tokens for user: %w", err)
	}
	return nil
}

func (s *TokenSessionStore) GetActiveByUUID(ctx context.Context, uuid string) (*models.AuthTokenSession, error) {
	session, err := s.getActive(ctx, "access_token_uuid = ?", uuid)
	if err != nil {
		return nil, err
	}
	return session, nil
}

func (s *TokenSessionStore) getActive(ctx context.Context, query string, value string) (*models.AuthTokenSession, error) {
	if s == nil || s.db == nil {
		return nil, ErrTokenSessionStoreUnavailable
	}
	var session models.AuthTokenSession
	err := s.db.WithContext(ctx).
		Where(query, value).
		Where("access_revoked_at IS NULL AND revoked_at IS NULL AND device_revoked_at IS NULL").
		Where("access_token_expires > ?", time.Now()).
		First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (s *TokenSessionStore) RevokeDevice(ctx context.Context, userID uint, deviceID string, revokedAt time.Time) ([]models.AuthTokenSession, error) {
	if s == nil || s.db == nil {
		return nil, ErrTokenSessionStoreUnavailable
	}
	var sessions []models.AuthTokenSession
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := lockUserForTokenMutation(tx, userID); err != nil {
			return err
		}
		if deviceID == "" {
			err := tx.Where(
				"user_id = ? AND device_id = '' AND revoked_at IS NULL", userID,
			).Find(&sessions).Error
			if err != nil {
				return err
			}
			return tx.Model(&models.AuthTokenSession{}).Where(
				"user_id = ? AND device_id = '' AND revoked_at IS NULL", userID,
			).Updates(map[string]interface{}{
				"access_revoked_at":  revokedAt,
				"refresh_revoked_at": revokedAt,
				"device_revoked_at":  revokedAt,
				"revoked_at":         revokedAt,
			}).Error
		}
		if err := tx.Where(
			"user_id = ? AND device_id = ? AND revoked_at IS NULL", userID, deviceID,
		).Find(&sessions).Error; err != nil {
			return err
		}
		return tx.Model(&models.AuthTokenSession{}).Where(
			"user_id = ? AND device_id = ? AND revoked_at IS NULL", userID, deviceID,
		).Updates(map[string]interface{}{
			"access_revoked_at":  revokedAt,
			"refresh_revoked_at": revokedAt,
			"device_revoked_at":  revokedAt,
			"revoked_at":         revokedAt,
		}).Error
	})
	if err != nil {
		return nil, fmt.Errorf("revoke auth tokens for device: %w", err)
	}
	return sessions, nil
}

func (s *TokenSessionStore) ListActiveDevices(ctx context.Context, userID uint) ([]models.AuthTokenSession, error) {
	if s == nil || s.db == nil {
		return nil, ErrTokenSessionStoreUnavailable
	}
	var sessions []models.AuthTokenSession
	err := s.db.WithContext(ctx).
		Where(
			"user_id = ? AND revoked_at IS NULL AND refresh_revoked_at IS NULL AND device_revoked_at IS NULL AND refresh_token_expires > ?",
			userID, time.Now(),
		).
		Order("created_at DESC").Find(&sessions).Error
	if err != nil {
		return nil, fmt.Errorf("list active auth sessions: %w", err)
	}
	return sessions, nil
}

func (s *TokenSessionStore) HasUserTokensRevokedAtOrAfter(ctx context.Context, userID uint, issuedAt time.Time) bool {
	if s == nil || s.db == nil || userID == 0 || issuedAt.IsZero() {
		return false
	}
	var count int64
	// A single-session logout or device revocation must not be interpreted as
	// a user-wide password reset/offboarding marker. Only explicit revoke-all
	// events invalidate every token issued at or before that event.
	err := s.db.WithContext(ctx).Model(&models.TokenRevocationLog{}).
		Where("user_id = ? AND revocation_type IN ? AND revoked_at > ?", userID, []string{string(RevokeAll), string(RevokeByUser)}, issuedAt).
		Limit(1).Count(&count).Error
	// Fail closed on database errors: a temporarily unavailable session store
	// must not resurrect tokens after a user/device revocation event.
	return err != nil || count > 0
}

func lockUserForTokenMutation(tx *gorm.DB, userID uint) error {
	// SQLite tests do not accept SELECT ... FOR UPDATE. On PostgreSQL and
	// MySQL, this user-row lock orders refresh rotation against user-wide and
	// device-wide revocation so neither can miss a concurrently added session.
	if tx.Dialector.Name() == "sqlite" {
		return nil
	}
	var lockedID uint
	return tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Model(&models.User{}).
		Select("id").
		Where("id = ?", userID).
		Scan(&lockedID).Error
}
