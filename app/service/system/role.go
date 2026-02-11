package system

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/seakee/go-api/app/model/system"
	"github.com/seakee/go-api/app/pkg/e"
	repo "github.com/seakee/go-api/app/repository/system"
	"github.com/sk-pkg/logger"
	"github.com/sk-pkg/redis"
	"gorm.io/gorm"
	"strings"
)

// RoleService defines the role service interface.
type RoleService interface {
	Create(ctx context.Context, role *system.Role) (errCode int, err error)
	Delete(ctx context.Context, id uint) (errCode int, err error)
	Detail(ctx context.Context, id uint) (role *system.Role, errCode int, err error)
	Paginate(ctx context.Context, role *system.Role, page, size int) (list []system.Role, total int64, err error)
	Update(ctx context.Context, role *system.Role) (errCode int, err error)
	List(ctx context.Context) (list []gin.H, err error)

	UpdatePermission(ctx context.Context, roleId uint, permissionIds []uint) (errCode int, err error)
	PermissionList(ctx context.Context, roleId uint) (list []uint, err error)
	PermissionMenuTree(ctx context.Context, roleId uint) (list []RoleMenuPermissionNode, errCode int, err error)
}

type RoleMenuPermissionNode struct {
	ID           uint                     `json:"id"`
	Name         string                   `json:"name"`
	Path         string                   `json:"path"`
	PermissionID uint                     `json:"permission_id"`
	ParentID     uint                     `json:"parent_id"`
	Icon         string                   `json:"icon"`
	Sort         int                      `json:"sort"`
	Checked      bool                     `json:"checked"`
	Children     []RoleMenuPermissionNode `json:"children,omitempty"`
}

type roleService struct {
	redis              *redis.Manager
	logger             *logger.Manager
	roleRepo           repo.RoleRepo
	roleUserRepo       repo.RoleUserRepo
	permissionRoleRepo repo.PermissionRoleRepo
	menuRepo           repo.MenuRepo
}

func (r roleService) PermissionMenuTree(ctx context.Context, roleId uint) (list []RoleMenuPermissionNode, errCode int, err error) {
	role, err := r.roleRepo.DetailByID(ctx, roleId)
	if role == nil {
		return nil, e.RoleNotFound, err
	}

	menus, err := r.menuRepo.List(ctx)
	if err != nil {
		return nil, e.ERROR, err
	}

	permissionIDs, err := r.permissionRoleRepo.ListByRoleID(ctx, roleId)
	if err != nil {
		return nil, e.ERROR, err
	}

	permissionSet := make(map[uint]struct{}, len(permissionIDs))
	for _, permissionID := range permissionIDs {
		permissionSet[permissionID] = struct{}{}
	}

	list = buildRoleMenuPermissionTree(menus, permissionSet)

	return list, e.SUCCESS, nil
}

func buildRoleMenuPermissionTree(menus system.MenuList, permissionSet map[uint]struct{}) []RoleMenuPermissionNode {
	list := make([]RoleMenuPermissionNode, len(menus))
	for i, menu := range menus {
		_, checked := permissionSet[menu.PermissionId]

		list[i] = RoleMenuPermissionNode{
			ID:           menu.ID,
			Name:         menu.Name,
			Path:         menu.Path,
			PermissionID: menu.PermissionId,
			ParentID:     menu.ParentId,
			Icon:         menu.Icon,
			Sort:         menu.Sort,
			Checked:      checked,
			Children:     buildRoleMenuPermissionTree(menu.Children, permissionSet),
		}
	}

	return list
}

func (r roleService) List(ctx context.Context) (list []gin.H, err error) {
	roles, err := r.roleRepo.List(ctx, &system.Role{})
	if err == nil && len(roles) > 0 {
		list = make([]gin.H, len(roles))
		for k, v := range roles {
			list[k] = make(gin.H)
			list[k]["id"] = v.ID
			list[k]["name"] = v.Name
			list[k]["description"] = v.Description
		}
	}

	return
}

func (r roleService) Paginate(ctx context.Context, role *system.Role, page, size int) (list []system.Role, total int64, err error) {
	total, err = r.roleRepo.Count(ctx, role)
	if err != nil {
		return nil, 0, err
	}

	list, err = r.roleRepo.Paginate(ctx, role, page, size)

	return
}

func (r roleService) Create(ctx context.Context, role *system.Role) (errCode int, err error) {
	errCode = e.RoleNameCantBeNull
	if role.Name != "" {
		var rol *system.Role

		errCode = e.RoleNameExists
		rol, err = r.roleRepo.Detail(ctx, &system.Role{Name: role.Name})
		if rol == nil {
			errCode = e.ERROR

			if role.Description == "" {
				role.Description = role.Name
			}

			_, err = r.roleRepo.Create(ctx, role)
			if err == nil {
				errCode = e.SUCCESS
			}
		}
	}

	return
}

func (r roleService) Delete(ctx context.Context, id uint) (int, error) {
	role, err := r.roleRepo.DetailByID(ctx, id)
	if role == nil {
		return e.RoleNotFound, err
	}

	roleName := strings.ToLower(role.Name)
	if roleName == "base" || roleName == "super_admin" {
		return e.RoleCanNotBeOperated, err
	}

	err = r.roleRepo.DeleteByID(ctx, id)
	if err == nil {
		err = r.roleUserRepo.DeleteByRoleID(ctx, id)
		if err == nil {
			err = r.permissionRoleRepo.DeleteByRoleID(ctx, id)
		}
	}

	if err == nil {
		return e.SUCCESS, nil
	}

	return e.ERROR, err
}

func (r roleService) Detail(ctx context.Context, id uint) (role *system.Role, errCode int, err error) {
	role, err = r.roleRepo.Detail(ctx, &system.Role{Model: gorm.Model{ID: id}})
	if err != nil {
		return nil, e.RoleNotFound, err
	}

	return role, e.SUCCESS, nil
}

func (r roleService) Update(ctx context.Context, role *system.Role) (errCode int, err error) {
	errCode = e.RoleNotFound

	rol, err := r.roleRepo.Detail(ctx, &system.Role{Model: gorm.Model{ID: role.ID}})
	if rol != nil {
		roleName := strings.ToLower(rol.Name)
		if roleName == "base" || roleName == "super_admin" {
			errCode = e.RoleCanNotBeOperated
			return
		}

		errCode = e.ERROR
		err = r.roleRepo.Update(ctx, role)
		if err == nil {
			errCode = e.SUCCESS
		}
	}

	return
}

// NewRoleService creates a new RoleService instance.
func NewRoleService(redis *redis.Manager, logger *logger.Manager, db *gorm.DB) RoleService {
	return &roleService{
		redis:              redis,
		logger:             logger,
		roleRepo:           repo.NewRoleRepo(db, redis, logger),
		roleUserRepo:       repo.NewRoleUserRepo(db, redis, logger),
		permissionRoleRepo: repo.NewPermissionRoleRepo(db, redis, logger),
		menuRepo:           repo.NewMenuRepo(db, redis, logger),
	}
}
