package system

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/seakee/go-api/app/model/system"
	"github.com/sk-pkg/logger"
	"github.com/sk-pkg/redis"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// UserIdentityRepo defines persistence for third-party user identities.
type UserIdentityRepo interface {
	DetailByProvider(ctx context.Context, provider, tenant, subject string) (*system.UserIdentity, error)
	DetailByIDAndUserID(ctx context.Context, id, userID uint) (*system.UserIdentity, error)
	ListByUserID(ctx context.Context, userID uint) ([]system.UserIdentity, error)
	Create(ctx context.Context, identity *system.UserIdentity) (*system.UserIdentity, error)
	UpdateLastLogin(ctx context.Context, id uint, loginAt time.Time) error
	DeleteByIDAndUserID(ctx context.Context, id, userID uint) error
	DeleteByUserID(ctx context.Context, userID uint) error
}

type userIdentityRepo struct {
	redis  *redis.Manager
	db     *gorm.DB
	logger *logger.Manager
}

func (u userIdentityRepo) DetailByProvider(ctx context.Context, provider, tenant, subject string) (*system.UserIdentity, error) {
	userIdentity := (&system.UserIdentity{}).Where(
		"provider = ? AND provider_tenant = ? AND provider_subject = ?",
		strings.TrimSpace(provider),
		strings.TrimSpace(tenant),
		strings.TrimSpace(subject),
	)

	return userIdentity.First(ctx, u.db)
}

func (u userIdentityRepo) DetailByIDAndUserID(ctx context.Context, id, userID uint) (*system.UserIdentity, error) {
	if id == 0 || userID == 0 {
		return nil, nil
	}

	userIdentity := (&system.UserIdentity{}).Where("id = ? AND user_id = ?", id, userID)

	return userIdentity.First(ctx, u.db)
}

func (u userIdentityRepo) ListByUserID(ctx context.Context, userID uint) ([]system.UserIdentity, error) {
	if userID == 0 {
		return make([]system.UserIdentity, 0), nil
	}

	userIdentity := (&system.UserIdentity{}).Where("user_id = ?", userID)

	list, err := userIdentity.FindWithSort(ctx, u.db, "id ASC")
	if err != nil {
		return nil, fmt.Errorf("list user identities failed: %w", err)
	}

	return list, nil
}

func (u userIdentityRepo) Create(ctx context.Context, identity *system.UserIdentity) (*system.UserIdentity, error) {
	if identity == nil {
		return nil, fmt.Errorf("user identity is nil")
	}

	if err := clearUserCache(ctx, u.redis); err != nil && u.logger != nil {
		u.logger.Error(ctx, "clear user cache failed", zap.String("action", "create user identity"), zap.Error(err))
	}

	if _, err := identity.Create(ctx, u.db); err != nil {
		return nil, fmt.Errorf("create user identity failed: %w", err)
	}

	return identity, nil
}

func (u userIdentityRepo) UpdateLastLogin(ctx context.Context, id uint, loginAt time.Time) error {
	if id == 0 {
		return fmt.Errorf("user identity id is empty")
	}

	userIdentity := (&system.UserIdentity{}).Where("id = ?", id)
	return userIdentity.Updates(ctx, u.db, map[string]interface{}{
		"last_login_at": loginAt,
	})
}

func (u userIdentityRepo) DeleteByIDAndUserID(ctx context.Context, id, userID uint) error {
	if id == 0 {
		return fmt.Errorf("user identity id is empty")
	}
	if userID == 0 {
		return fmt.Errorf("user id is empty")
	}

	if err := clearUserCache(ctx, u.redis); err != nil && u.logger != nil {
		u.logger.Error(ctx, "clear user cache failed", zap.String("action", "delete user identity"), zap.Error(err))
	}

	userIdentity := (&system.UserIdentity{}).Where("id = ? AND user_id = ?", id, userID)
	rowsAffected, err := userIdentity.UnscopedDelete(ctx, u.db)
	if err != nil {
		return fmt.Errorf("delete user identity failed: %w", err)
	}
	if rowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (u userIdentityRepo) DeleteByUserID(ctx context.Context, userID uint) error {
	if userID == 0 {
		return fmt.Errorf("user id is empty")
	}

	if err := clearUserCache(ctx, u.redis); err != nil && u.logger != nil {
		u.logger.Error(ctx, "clear user cache failed", zap.String("action", "delete user identities"), zap.Error(err))
	}

	return u.db.WithContext(ctx).Unscoped().Where("user_id = ?", userID).Delete(&system.UserIdentity{}).Error
}

// NewUserIdentityRepo creates a new UserIdentityRepo instance.
func NewUserIdentityRepo(db *gorm.DB, redis *redis.Manager, logger *logger.Manager) UserIdentityRepo {
	return &userIdentityRepo{db: db, redis: redis, logger: logger}
}
