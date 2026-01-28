package system

import (
	"context"
	"errors"
	"fmt"
	"github.com/seakee/go-api/app/model/system"
	"github.com/seakee/go-api/app/model/system/role"
	"github.com/sk-pkg/logger"
	"github.com/sk-pkg/redis"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// RoleUserRepo defines the role-user repository interface.
type RoleUserRepo interface {
	i()
	ListByUserID(ctx context.Context, userID uint) (roles []uint, err error)
	UpdateUserRole(ctx context.Context, userID uint, roles []uint) error
	DeleteByRoleID(ctx context.Context, roleID uint) error
}

type roleUserRepo struct {
	redis  *redis.Manager
	db     *gorm.DB
	logger *logger.Manager
}

func (rup roleUserRepo) DeleteByRoleID(ctx context.Context, roleID uint) error {
	err := clearUserCache(ctx, rup.redis)
	if err != nil {
		rup.logger.Error(ctx, "clear user cache failed", zap.String("action", "delete role user"), zap.Error(err))
	}

	ru := &role.User{RoleId: roleID}
	return ru.Delete(ctx, rup.db)
}

func (rup roleUserRepo) UpdateUserRole(ctx context.Context, userID uint, roles []uint) error {
	// Clear cache
	if err := clearUserCache(ctx, rup.redis); err != nil {
		rup.logger.Error(ctx, "clear user cache failed", zap.String("action", "update role user"), zap.Error(err))
	}

	// Deduplicate and add base role
	roleSet := make(map[uint]struct{}, len(roles)+1)
	for _, roleID := range roles {
		roleSet[roleID] = struct{}{}
	}

	// Add base role if it exists
	if baseRole, err := rup.getBaseRole(ctx); err == nil {
		roleSet[baseRole.ID] = struct{}{}
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("failed to get base role: %w", err)
	}

	// Build role-user relationships
	roleUsers := make([]role.User, 0, len(roleSet))
	for roleID := range roleSet {
		roleUsers = append(roleUsers, role.User{
			UserId: userID,
			RoleId: roleID,
		})
	}

	// Execute transaction
	return rup.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Delete old role relationships
		if err := tx.Where("user_id = ?", userID).Delete(&role.User{}).Error; err != nil {
			return fmt.Errorf("failed to delete old user roles: %w", err)
		}

		// Create new role relationships
		if err := tx.Create(&roleUsers).Error; err != nil {
			return fmt.Errorf("failed to create new user roles: %w", err)
		}

		return nil
	})
}

// getBaseRole retrieves the base role.
func (rup roleUserRepo) getBaseRole(ctx context.Context) (*system.Role, error) {
	var baseRole system.Role
	err := rup.db.WithContext(ctx).Where("name = ?", "base").First(&baseRole).Error
	if err != nil {
		return nil, err
	}

	return &baseRole, nil
}

func (rup roleUserRepo) ListByUserID(ctx context.Context, userID uint) (roles []uint, err error) {
	ru := &role.User{UserId: userID}
	list, err := ru.List(ctx, rup.db)
	if err == nil {
		roles = make([]uint, len(list))
		for k, v := range list {
			roles[k] = v.RoleId
		}
	}

	return
}

func (rup roleUserRepo) i() {
	panic("implement me")
}

// NewRoleUserRepo creates a new RoleUserRepo instance.
func NewRoleUserRepo(db *gorm.DB, redis *redis.Manager, logger *logger.Manager) RoleUserRepo {
	return &roleUserRepo{redis: redis, db: db, logger: logger}
}
