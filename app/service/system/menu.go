package system

import (
	"context"
	"github.com/seakee/go-api/app/model/system"
	"github.com/seakee/go-api/app/pkg/e"
	repo "github.com/seakee/go-api/app/repository/system"
	"github.com/sk-pkg/logger"
	"github.com/sk-pkg/redis"
	"gorm.io/gorm"
)

// MenuService defines the menu service interface.
type MenuService interface {
	Create(ctx context.Context, menu *system.Menu) (errCode int, err error)
	Delete(ctx context.Context, id uint) (errCode int, err error)
	Detail(ctx context.Context, id uint) (menu *system.Menu, errCode int, err error)
	List(ctx context.Context) (list system.MenuList, err error)
	Update(ctx context.Context, menu *system.Menu) (errCode int, err error)
}

type menuService struct {
	redis    *redis.Manager
	logger   *logger.Manager
	db       *gorm.DB
	menuRepo repo.MenuRepo
	authRepo repo.AuthRepo
}

func (m menuService) Create(ctx context.Context, menu *system.Menu) (errCode int, err error) {
	var getMenu *system.Menu

	getMenu, err = m.menuRepo.Detail(ctx, &system.Menu{Name: menu.Name})
	if getMenu != nil {
		return e.MenuNameExists, nil
	}

	if menu.ParentId != 0 {
		var parentMenu *system.Menu
		parentMenu, err = m.menuRepo.Detail(ctx, &system.Menu{Model: gorm.Model{ID: menu.ParentId}})
		if parentMenu == nil {
			return e.MenuNotFound, nil
		}
	}

	_, err = m.menuRepo.Create(ctx, menu)
	if err != nil {
		return e.ERROR, err
	}

	return e.SUCCESS, nil
}

func (m menuService) Delete(ctx context.Context, id uint) (errCode int, err error) {
	var count int64

	count, err = m.menuRepo.Count(ctx, &system.Menu{ParentId: id})
	if count > 0 {
		return e.MenuHasSubMenu, nil
	}

	err = m.menuRepo.DeleteByID(ctx, id)
	if err != nil {
		return e.ERROR, err
	}

	return e.SUCCESS, nil
}

func (m menuService) Detail(ctx context.Context, id uint) (menu *system.Menu, errCode int, err error) {
	var detail *system.Menu

	detail, err = m.menuRepo.DetailByID(ctx, id)
	if detail == nil {
		return nil, e.MenuNotFound, nil
	}

	return detail, e.SUCCESS, nil
}

func (m menuService) List(ctx context.Context) (list system.MenuList, err error) {
	return m.menuRepo.List(ctx)
}

func (m menuService) Update(ctx context.Context, menu *system.Menu) (errCode int, err error) {
	var detail *system.Menu

	detail, err = m.menuRepo.DetailByID(ctx, menu.ID)
	if detail == nil {
		return e.MenuNotFound, nil
	}

	err = m.menuRepo.Update(ctx, menu)
	if err != nil {
		return e.ERROR, err
	}

	return e.SUCCESS, nil
}

// NewMenuService creates a new MenuService instance.
func NewMenuService(redis *redis.Manager, logger *logger.Manager, db *gorm.DB) MenuService {
	return &menuService{
		redis:    redis,
		logger:   logger,
		db:       db,
		menuRepo: repo.NewMenuRepo(db, redis, logger),
		authRepo: repo.NewAuthRepo(db, redis, logger),
	}
}
