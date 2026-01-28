package system

import (
	"context"
	"fmt"
	"github.com/seakee/go-api/app/model/system/permission"
	"github.com/sk-pkg/logger"
	"github.com/sk-pkg/redis"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// PermissionRoleRepo defines the permission-role repository interface.
type PermissionRoleRepo interface {
	i()
	ListByRoleID(ctx context.Context, roleID uint) (permission []uint, err error)
	UpdateRolePermission(ctx context.Context, roleID uint, permissions []uint) error
	DeleteByPermID(ctx context.Context, permID uint) error
	DeleteByRoleID(ctx context.Context, roleID uint) error
}

type permissionRoleRepo struct {
	redis  *redis.Manager
	db     *gorm.DB
	logger *logger.Manager
}

func (pr permissionRoleRepo) DeleteByRoleID(ctx context.Context, roleID uint) error {
	err := clearUserCache(ctx, pr.redis)
	if err != nil {
		pr.logger.Error(ctx, "clear user cache failed", zap.String("action", "delete permission role"), zap.Error(err))
	}

	prm := &permission.Role{RoleId: roleID}
	return prm.Delete(ctx, pr.db)
}

func (pr permissionRoleRepo) DeleteByPermID(ctx context.Context, permID uint) error {
	err := clearUserCache(ctx, pr.redis)
	if err != nil {
		pr.logger.Error(ctx, "clear user cache failed", zap.String("action", "delete permission role"), zap.Error(err))
	}

	prm := &permission.Role{PermissionId: permID}
	return prm.Delete(ctx, pr.db)
}

func (pr permissionRoleRepo) i() {
	panic("implement me")
}

func (pr permissionRoleRepo) UpdateRolePermission(ctx context.Context, roleID uint, permIDs []uint) error {
	// Clear user cache
	if err := clearUserCache(ctx, pr.redis); err != nil {
		pr.logger.Error(ctx, "clear user cache failed", zap.String("action", "update permission role"), zap.Error(err))
	}

	// Use map to optimize existence checking
	permIDSet := make(map[uint]struct{}, len(permIDs))
	for _, id := range permIDs {
		permIDSet[id] = struct{}{}
	}

	return pr.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Get existing permission IDs
		var existingIDs []uint
		if err := tx.Model(&permission.Role{}).
			Where("role_id = ?", roleID).
			Pluck("permission_id", &existingIDs).Error; err != nil {
			return fmt.Errorf("failed to fetch existing permissions: %w", err)
		}

		// Calculate permissions to delete and add
		toDelete := make([]uint, 0)
		toAdd := make([]permission.Role, 0)

		for _, existID := range existingIDs {
			if _, exists := permIDSet[existID]; !exists {
				toDelete = append(toDelete, existID)
			}
		}

		for permID := range permIDSet {
			var found bool
			for _, existID := range existingIDs {
				if permID == existID {
					found = true
					break
				}
			}
			if !found {
				toAdd = append(toAdd, permission.Role{
					PermissionId: permID,
					RoleId:       roleID,
				})
			}
		}

		// Delete unnecessary permissions
		if len(toDelete) > 0 {
			if err := tx.Where("role_id = ? AND permission_id IN (?)", roleID, toDelete).
				Delete(&permission.Role{}).Error; err != nil {
				return fmt.Errorf("failed to delete permissions: %w", err)
			}
		}

		// Add new permissions
		if len(toAdd) > 0 {
			if err := tx.Create(&toAdd).Error; err != nil {
				return fmt.Errorf("failed to add new permissions: %w", err)
			}
		}

		return nil
	})
}

func (pr permissionRoleRepo) ListByRoleID(ctx context.Context, roleID uint) (permissionIDs []uint, err error) {
	prm := &permission.Role{RoleId: roleID}
	list, err := prm.List(ctx, pr.db)
	if err == nil {
		permissionIDs = make([]uint, len(list))
		for k, v := range list {
			permissionIDs[k] = v.PermissionId
		}
	}
	return
}

// NewPermissionRoleRepo creates a new PermissionRoleRepo instance.
func NewPermissionRoleRepo(db *gorm.DB, redis *redis.Manager, logger *logger.Manager) PermissionRoleRepo {
	return &permissionRoleRepo{redis: redis, db: db, logger: logger}
}
