package system

import (
	"context"
	"fmt"
	"github.com/seakee/go-api/app/model/system"
	"github.com/seakee/go-api/app/model/system/permission"
	"github.com/seakee/go-api/app/model/system/role"
	"github.com/sk-pkg/logger"
	"github.com/sk-pkg/redis"
	"gorm.io/gorm"
)

// MenuRepo defines the menu repository interface.
type MenuRepo interface {
	UserMenuList(ctx context.Context, userID uint, isSuperAdmin bool) (system.MenuList, error)
	List(ctx context.Context) (system.MenuList, error)
	DetailByID(ctx context.Context, id uint) (*system.Menu, error)
	Detail(ctx context.Context, menu *system.Menu) (*system.Menu, error)
	Create(ctx context.Context, menu *system.Menu) (uint, error)
	DeleteByID(ctx context.Context, id uint) error
	Update(ctx context.Context, menu *system.Menu) error
	Count(ctx context.Context, menu *system.Menu) (int64, error)
}

type menuRepo struct {
	db     *gorm.DB
	redis  *redis.Manager
	logger *logger.Manager
}

func (m menuRepo) Count(ctx context.Context, menu *system.Menu) (int64, error) {
	return menu.Count(ctx, m.db)
}

func (m menuRepo) UserMenuList(ctx context.Context, userID uint, isSuperAdmin bool) (system.MenuList, error) {
	if isSuperAdmin {
		return m.getAllMenus(ctx)
	}

	return m.getUserSpecificMenus(ctx, userID)
}

func (m menuRepo) getAllMenus(ctx context.Context) (system.MenuList, error) {
	var menuList system.MenuList
	if err := m.db.WithContext(ctx).Find(&menuList).Error; err != nil {
		return nil, fmt.Errorf("failed to get all menus: %w", err)
	}

	return menuList.GenTree(), nil
}

func (m menuRepo) getUserSpecificMenus(ctx context.Context, userID uint) (system.MenuList, error) {
	roleIDs, err := m.getUserRoleIDs(ctx, userID)
	if err != nil {
		return nil, err
	}

	permissionIDs, err := m.getRolePermissionIDs(ctx, roleIDs)
	if err != nil {
		return nil, err
	}

	userMenus, err := m.getUserMenusByPermissions(ctx, permissionIDs)
	if err != nil {
		return nil, err
	}

	allMenuIDs, err := m.getAllMenuIDs(ctx, userMenus)
	if err != nil {
		return nil, err
	}

	return m.getMenusByIDs(ctx, allMenuIDs)
}

func (m menuRepo) getUserRoleIDs(ctx context.Context, userID uint) ([]uint, error) {
	var roleList []role.User
	var err error

	roleUser := role.User{UserId: userID}
	roleList, err = roleUser.List(ctx, m.db)
	if err != nil {
		return nil, err
	}

	roleIDs := make([]uint, len(roleList))
	for i, r := range roleList {
		roleIDs[i] = r.RoleId
	}

	return roleIDs, nil
}

func (m menuRepo) getRolePermissionIDs(ctx context.Context, roleIDs []uint) ([]uint, error) {
	var permissionsOfRole []permission.Role
	if err := m.db.WithContext(ctx).Model(&permission.Role{}).Select("permission_id").Where("role_id IN ?", roleIDs).Find(&permissionsOfRole).Error; err != nil {
		return nil, fmt.Errorf("failed to get role permissions: %w", err)
	}

	permissionIDs := make([]uint, len(permissionsOfRole))
	for i, p := range permissionsOfRole {
		permissionIDs[i] = p.PermissionId
	}

	return permissionIDs, nil
}

func (m menuRepo) getUserMenusByPermissions(ctx context.Context, permissionIDs []uint) (system.MenuList, error) {
	var userMenus system.MenuList
	if err := m.db.WithContext(ctx).Where("permission_id IN ?", permissionIDs).Find(&userMenus).Error; err != nil {
		return nil, fmt.Errorf("failed to get user menus: %w", err)
	}

	return userMenus, nil
}

func (m menuRepo) getAllMenuIDs(ctx context.Context, userMenus system.MenuList) ([]uint, error) {
	var allMenus system.MenuList
	if err := m.db.WithContext(ctx).Find(&allMenus).Error; err != nil {
		return nil, fmt.Errorf("failed to get all menus: %w", err)
	}

	menuIdsOfRole := make(map[uint]struct{})
	for _, menu := range userMenus {
		allMenus.AllUserMenuIds(menu.ID, menuIdsOfRole)
	}

	menuIDs := make([]uint, 0, len(menuIdsOfRole))
	for id := range menuIdsOfRole {
		menuIDs = append(menuIDs, id)
	}

	return menuIDs, nil
}

func (m menuRepo) getMenusByIDs(ctx context.Context, menuIDs []uint) (system.MenuList, error) {
	var userMenus system.MenuList
	if err := m.db.WithContext(ctx).Where("id IN ?", menuIDs).Find(&userMenus).Error; err != nil {
		return nil, fmt.Errorf("failed to get menus by IDs: %w", err)
	}

	return userMenus.GenTree(), nil
}

func (m menuRepo) List(ctx context.Context) (system.MenuList, error) {
	var menuList system.MenuList
	var err error

	menu := system.Menu{}

	menuList, err = menu.List(ctx, m.db)
	if err != nil {
		return nil, fmt.Errorf("failed to get menu list: %w", err)
	}

	return menuList.GenTree(), nil
}

func (m menuRepo) DetailByID(ctx context.Context, id uint) (*system.Menu, error) {
	menu := &system.Menu{Model: gorm.Model{ID: id}}

	return m.Detail(ctx, menu)
}

func (m menuRepo) Detail(ctx context.Context, menu *system.Menu) (*system.Menu, error) {
	return menu.First(ctx, m.db)
}

func (m menuRepo) Create(ctx context.Context, menu *system.Menu) (uint, error) {
	err := m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		perm := system.Permission{
			Name:        menu.Name,
			Type:        "menu",
			Description: menu.Name,
			Group:       "sys-menu",
		}

		if err := tx.Create(&perm).Error; err != nil {
			return fmt.Errorf("failed to create permission: %w", err)
		}

		menu.PermissionId = perm.ID
		if err := tx.Create(menu).Error; err != nil {
			return fmt.Errorf("failed to create menu: %w", err)
		}

		return nil
	})

	return menu.ID, err
}

func (m menuRepo) DeleteByID(ctx context.Context, id uint) error {
	return m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var menu system.Menu
		if err := tx.First(&menu, id).Error; err != nil {
			return fmt.Errorf("failed to find menu: %w", err)
		}

		if err := tx.Delete(&system.Menu{}, id).Error; err != nil {
			return fmt.Errorf("failed to delete menu: %w", err)
		}

		if err := tx.Delete(&system.Permission{}, menu.PermissionId).Error; err != nil {
			return fmt.Errorf("failed to delete permission: %w", err)
		}

		return nil
	})
}

func (m menuRepo) Update(ctx context.Context, menu *system.Menu) error {
	return m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(menu).Select("name", "path", "icon", "sort", "title").Updates(menu).Error; err != nil {
			return fmt.Errorf("failed to update menu: %w", err)
		}

		if menu.Name != "" {
			if err := tx.Model(&system.Permission{}).Where("id = ?", menu.PermissionId).Update("name", menu.Name).Error; err != nil {
				return fmt.Errorf("failed to update permission: %w", err)
			}
		}

		return nil
	})
}

// NewMenuRepo creates a new MenuRepo instance.
func NewMenuRepo(db *gorm.DB, redis *redis.Manager, logger *logger.Manager) MenuRepo {
	return &menuRepo{
		db:     db,
		redis:  redis,
		logger: logger,
	}
}
