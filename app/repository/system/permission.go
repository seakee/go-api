package system

import (
	"context"
	"fmt"
	"github.com/seakee/go-api/app/model/system"
	"github.com/sk-pkg/logger"
	"github.com/sk-pkg/redis"
	"github.com/sk-pkg/util"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// PermissionRepo defines the permission repository interface.
type PermissionRepo interface {
	i()
	HasHash(ctx context.Context, hash string) bool
	Paginate(ctx context.Context, permission *system.Permission, page, pageSize int) ([]system.Permission, error)
	ListByType(ctx context.Context, permType string) ([]system.Permission, error)
	DetailByID(ctx context.Context, id uint) (*system.Permission, error)
	Detail(ctx context.Context, permission *system.Permission) (*system.Permission, error)
	Create(ctx context.Context, permission *system.Permission) (uint, error)
	DeleteByID(ctx context.Context, id uint) error
	Update(ctx context.Context, permission *system.Permission) error
	Count(ctx context.Context, permission *system.Permission) (int64, error)
}

type permissionRepo struct {
	redis  *redis.Manager
	db     *gorm.DB
	logger *logger.Manager
}

func (p permissionRepo) Count(ctx context.Context, permission *system.Permission) (int64, error) {
	return permission.Count(ctx, p.db)
}

func (p permissionRepo) Paginate(ctx context.Context, permission *system.Permission, page, pageSize int) ([]system.Permission, error) {
	if page == 0 {
		page = 1
	}

	switch {
	case pageSize > 100:
		pageSize = 100
	case pageSize <= 0:
		pageSize = 10
	}

	return permission.FindWithPagination(ctx, p.db, page, pageSize)
}

func (p permissionRepo) ListByType(ctx context.Context, permType string) ([]system.Permission, error) {
	permission := &system.Permission{Type: permType}
	return permission.List(ctx, p.db)
}

func (p permissionRepo) DetailByID(ctx context.Context, id uint) (*system.Permission, error) {
	permission := &system.Permission{Model: gorm.Model{ID: id}}

	return permission.First(ctx, p.db)
}

func (p permissionRepo) Detail(ctx context.Context, permission *system.Permission) (*system.Permission, error) {
	return permission.First(ctx, p.db)
}

func (p permissionRepo) Create(ctx context.Context, permission *system.Permission) (id uint, err error) {
	err = clearUserCache(ctx, p.redis)
	if err != nil {
		p.logger.Error(ctx, "clear user cache failed", zap.String("action", "create permission"), zap.Error(err))
	}

	return permission.Create(ctx, p.db)
}

func (p permissionRepo) DeleteByID(ctx context.Context, id uint) error {
	err := clearUserCache(ctx, p.redis)
	if err != nil {
		p.logger.Error(ctx, "clear user cache failed", zap.String("action", "delete permission"), zap.Error(err))
	}

	permission := &system.Permission{Model: gorm.Model{ID: id}}

	return permission.Delete(ctx, p.db)
}

func (p permissionRepo) Update(ctx context.Context, permission *system.Permission) error {
	if permission.ID <= 0 {
		return fmt.Errorf("permission id is empty")
	}

	err := clearUserCache(ctx, p.redis)
	if err != nil {
		p.logger.Error(ctx, "clear user cache failed", zap.String("action", "update permission"), zap.Error(err))
	}

	data := make(map[string]interface{})
	if permission.Name != "" {
		data["name"] = permission.Name
	}

	if permission.Type != "" {
		data["type"] = permission.Type
	}

	if permission.Group != "" {
		data["group"] = permission.Group
	}

	if permission.Method != "" {
		data["method"] = permission.Method
	}

	if permission.Path != "" {
		data["path"] = permission.Path
	}

	if permission.Description != "" {
		data["description"] = permission.Description
	}

	updatePerm := &system.Permission{Model: gorm.Model{ID: permission.ID}}

	return updatePerm.Updates(ctx, p.db, data)
}

func (p permissionRepo) HasHash(ctx context.Context, hash string) bool {
	var list []system.Permission

	cacheKey := util.SpliceStr(userCachePrefix, ":permissions:list")

	err := p.redis.GetJSON(cacheKey, &list)
	if err != nil {
		perm := &system.Permission{}
		list, err = perm.List(ctx, p.db)
		if err != nil {
			return false
		}

		err = p.redis.SetJSON(cacheKey, list, 7200)
		if err != nil {
			p.logger.Error(ctx, "set permission list cache failed", zap.String("action", "has hash"), zap.Error(err))
		}
	}

	if err == nil && len(list) > 0 {
		for _, v := range list {
			if util.MD5(v.Method+v.Path) == hash {
				return true
			}
		}
	}

	return false
}

func (p permissionRepo) i() {
	panic("implement me")
}

// NewPermissionRepo creates a new PermissionRepo instance.
func NewPermissionRepo(db *gorm.DB, redis *redis.Manager, logger *logger.Manager) PermissionRepo {
	return &permissionRepo{redis: redis, db: db, logger: logger}
}
