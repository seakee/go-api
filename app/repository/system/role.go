package system

import (
	"context"
	"fmt"
	"github.com/seakee/go-api/app/model/system"
	"github.com/sk-pkg/logger"
	"github.com/sk-pkg/redis"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// RoleRepo defines the role repository interface.
type RoleRepo interface {
	i()
	Paginate(ctx context.Context, role *system.Role, page, pageSize int) ([]system.Role, error)
	List(ctx context.Context, role *system.Role) (list []system.Role, err error)
	DetailByID(ctx context.Context, id uint) (role *system.Role, err error)
	Detail(ctx context.Context, role *system.Role) (*system.Role, error)
	Create(ctx context.Context, role *system.Role) (id uint, err error)
	DeleteByID(ctx context.Context, id uint) error
	Update(ctx context.Context, role *system.Role) error
	Count(ctx context.Context, role *system.Role) (int64, error)
}

type roleRepo struct {
	redis  *redis.Manager
	db     *gorm.DB
	logger *logger.Manager
}

func (r roleRepo) Count(ctx context.Context, role *system.Role) (int64, error) {
	return role.Count(ctx, r.db)
}

func (r roleRepo) Update(ctx context.Context, role *system.Role) error {
	if role.ID <= 0 {
		return fmt.Errorf("role id is empty")
	}

	err := clearUserCache(ctx, r.redis)
	if err != nil {
		r.logger.Error(ctx, "clear user cache failed", zap.String("action", "update role"), zap.Error(err))
	}

	data := make(map[string]interface{})
	if role.Name != "" {
		data["name"] = role.Name
	}

	if role.Description != "" {
		data["description"] = role.Description
	}

	updateRole := &system.Role{Model: gorm.Model{ID: role.ID}}

	return updateRole.Updates(ctx, r.db, data)
}

func (r roleRepo) DeleteByID(ctx context.Context, id uint) error {
	err := clearUserCache(ctx, r.redis)
	if err != nil {
		r.logger.Error(ctx, "clear user cache failed", zap.String("action", "delete role"), zap.Error(err))
	}

	role := &system.Role{Model: gorm.Model{ID: id}}

	return role.Delete(ctx, r.db)
}

func (r roleRepo) Create(ctx context.Context, role *system.Role) (id uint, err error) {
	err = clearUserCache(ctx, r.redis)
	if err != nil {
		r.logger.Error(ctx, "clear user cache failed", zap.String("action", "create role"), zap.Error(err))
	}

	return role.Create(ctx, r.db)
}

func (r roleRepo) Detail(ctx context.Context, role *system.Role) (*system.Role, error) {
	return role.First(ctx, r.db)
}

func (r roleRepo) DetailByID(ctx context.Context, id uint) (*system.Role, error) {
	role := &system.Role{Model: gorm.Model{ID: id}}

	return role.First(ctx, r.db)
}

func (r roleRepo) i() {
	panic("implement me")
}

func (r roleRepo) List(ctx context.Context, role *system.Role) ([]system.Role, error) {
	return role.List(ctx, r.db)
}

func (r roleRepo) Paginate(ctx context.Context, role *system.Role, page, pageSize int) ([]system.Role, error) {
	if page == 0 {
		page = 1
	}

	switch {
	case pageSize > 100:
		pageSize = 100
	case pageSize <= 0:
		pageSize = 10
	}

	return role.FindWithPagination(ctx, r.db, page, pageSize)
}

// NewRoleRepo creates a new RoleRepo instance.
func NewRoleRepo(db *gorm.DB, redis *redis.Manager, logger *logger.Manager) RoleRepo {
	return &roleRepo{redis: redis, db: db, logger: logger}
}
