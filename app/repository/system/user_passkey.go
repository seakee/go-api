package system

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/seakee/go-api/app/model/system"
	"github.com/sk-pkg/logger"
	"github.com/sk-pkg/redis"
	"gorm.io/gorm"
)

// UserPasskeyRepo defines persistence operations for passkey credentials.
type UserPasskeyRepo interface {
	DetailByCredentialID(ctx context.Context, credentialID string) (*system.UserPasskey, error)
	DetailByIDAndUserID(ctx context.Context, id, userID uint) (*system.UserPasskey, error)
	ListByUserID(ctx context.Context, userID uint) ([]system.UserPasskey, error)
	CountByUserID(ctx context.Context, userID uint) (int64, error)
	CountByUserIDs(ctx context.Context, userIDs []uint) (map[uint]int64, error)
	Create(ctx context.Context, passkey *system.UserPasskey) (*system.UserPasskey, error)
	UpdateSignCount(ctx context.Context, id uint, signCount uint32) error
	UpdateLastUsedAt(ctx context.Context, id uint, usedAt time.Time) error
	DeleteByIDAndUserID(ctx context.Context, id, userID uint) error
	DeleteByID(ctx context.Context, id uint) error
	DeleteByUserID(ctx context.Context, userID uint) error
}

type userPasskeyRepo struct {
	redis  *redis.Manager
	db     *gorm.DB
	logger *logger.Manager
}

func (u userPasskeyRepo) DetailByCredentialID(ctx context.Context, credentialID string) (*system.UserPasskey, error) {
	credentialID = strings.TrimSpace(credentialID)
	if credentialID == "" {
		return nil, nil
	}

	var passkey system.UserPasskey
	if err := u.db.WithContext(ctx).Where("credential_id = ?", credentialID).First(&passkey).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("detail passkey by credential id failed: %w", err)
	}

	return &passkey, nil
}

func (u userPasskeyRepo) DetailByIDAndUserID(ctx context.Context, id, userID uint) (*system.UserPasskey, error) {
	if id == 0 || userID == 0 {
		return nil, nil
	}

	var passkey system.UserPasskey
	if err := u.db.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).First(&passkey).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("detail passkey by id and user id failed: %w", err)
	}

	return &passkey, nil
}

func (u userPasskeyRepo) ListByUserID(ctx context.Context, userID uint) ([]system.UserPasskey, error) {
	if userID == 0 {
		return make([]system.UserPasskey, 0), nil
	}

	list := make([]system.UserPasskey, 0)
	if err := u.db.WithContext(ctx).Where("user_id = ?", userID).Order("id ASC").Find(&list).Error; err != nil {
		return nil, fmt.Errorf("list passkeys by user id failed: %w", err)
	}

	return list, nil
}

func (u userPasskeyRepo) CountByUserID(ctx context.Context, userID uint) (int64, error) {
	if userID == 0 {
		return 0, nil
	}

	var count int64
	if err := u.db.WithContext(ctx).Model(&system.UserPasskey{}).Where("user_id = ?", userID).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("count passkeys by user id failed: %w", err)
	}

	return count, nil
}

func (u userPasskeyRepo) CountByUserIDs(ctx context.Context, userIDs []uint) (map[uint]int64, error) {
	result := make(map[uint]int64, len(userIDs))
	if len(userIDs) == 0 {
		return result, nil
	}
	for _, userID := range userIDs {
		result[userID] = 0
	}

	type countRow struct {
		UserID uint  `gorm:"column:user_id"`
		Count  int64 `gorm:"column:count"`
	}
	rows := make([]countRow, 0)
	if err := u.db.WithContext(ctx).
		Model(&system.UserPasskey{}).
		Select("user_id, COUNT(*) AS count").
		Where("user_id IN ?", userIDs).
		Group("user_id").
		Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("count passkeys by user ids failed: %w", err)
	}

	for _, row := range rows {
		result[row.UserID] = row.Count
	}

	return result, nil
}

func (u userPasskeyRepo) Create(ctx context.Context, passkey *system.UserPasskey) (*system.UserPasskey, error) {
	if passkey == nil {
		return nil, fmt.Errorf("passkey is nil")
	}

	if err := u.db.WithContext(ctx).Create(passkey).Error; err != nil {
		return nil, fmt.Errorf("create passkey failed: %w", err)
	}

	return passkey, nil
}

func (u userPasskeyRepo) UpdateSignCount(ctx context.Context, id uint, signCount uint32) error {
	if id == 0 {
		return fmt.Errorf("passkey id is empty")
	}

	return u.db.WithContext(ctx).Model(&system.UserPasskey{}).Where("id = ?", id).Updates(map[string]interface{}{
		"sign_count": signCount,
	}).Error
}

func (u userPasskeyRepo) UpdateLastUsedAt(ctx context.Context, id uint, usedAt time.Time) error {
	if id == 0 {
		return fmt.Errorf("passkey id is empty")
	}

	return u.db.WithContext(ctx).Model(&system.UserPasskey{}).Where("id = ?", id).Updates(map[string]interface{}{
		"last_used_at": usedAt,
	}).Error
}

func (u userPasskeyRepo) DeleteByIDAndUserID(ctx context.Context, id, userID uint) error {
	if id == 0 {
		return fmt.Errorf("passkey id is empty")
	}
	if userID == 0 {
		return fmt.Errorf("user id is empty")
	}

	result := u.db.WithContext(ctx).Unscoped().Where("id = ? AND user_id = ?", id, userID).Delete(&system.UserPasskey{})
	if result.Error != nil {
		return fmt.Errorf("delete passkey by id and user id failed: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (u userPasskeyRepo) DeleteByID(ctx context.Context, id uint) error {
	if id == 0 {
		return fmt.Errorf("passkey id is empty")
	}

	result := u.db.WithContext(ctx).Unscoped().Where("id = ?", id).Delete(&system.UserPasskey{})
	if result.Error != nil {
		return fmt.Errorf("delete passkey by id failed: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (u userPasskeyRepo) DeleteByUserID(ctx context.Context, userID uint) error {
	if userID == 0 {
		return fmt.Errorf("user id is empty")
	}

	return u.db.WithContext(ctx).Unscoped().Where("user_id = ?", userID).Delete(&system.UserPasskey{}).Error
}

// NewUserPasskeyRepo creates a new UserPasskeyRepo instance.
func NewUserPasskeyRepo(db *gorm.DB, redis *redis.Manager, logger *logger.Manager) UserPasskeyRepo {
	return &userPasskeyRepo{db: db, redis: redis, logger: logger}
}
