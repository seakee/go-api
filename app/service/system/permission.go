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

// PermissionService defines the permission service interface.
type PermissionService interface {
	Create(ctx context.Context, permission *system.Permission) (errCode int, err error)
	Delete(ctx context.Context, id uint) error
	Detail(ctx context.Context, id uint) (permission *system.Permission, errCode int, err error)
	ListInGroup(ctx context.Context, permType string) (list map[string][]system.Permission, err error)
	Paginate(ctx context.Context, permission *system.Permission, page, size int) (list []system.Permission, total int64, err error)
	Update(ctx context.Context, permission *system.Permission) (errCode int, err error)
	Available(ctx context.Context) (map[string][]string, error)
}

type permissionService struct {
	redis        *redis.Manager
	logger       *logger.Manager
	routes       func() gin.RoutesInfo
	permRepo     repo.PermissionRepo
	permRoleRepo repo.PermissionRoleRepo
}

func (p permissionService) Paginate(ctx context.Context, permission *system.Permission, page, size int) (list []system.Permission, total int64, err error) {
	total, err = p.permRepo.Count(ctx, permission)
	if err != nil {
		return nil, 0, err
	}

	list, err = p.permRepo.Paginate(ctx, permission, page, size)

	return
}

// Available retrieves available API routes that have not been assigned permissions.
// Returns a map where keys are HTTP methods and values are slices of paths.
func (p permissionService) Available(ctx context.Context) (map[string][]string, error) {
	// Get all existing API permissions from the repository
	perms, err := p.permRepo.ListByType(ctx, "api")
	if err != nil {
		return nil, err // Return error if unable to get permissions
	}

	// Initialize a map to store available routes
	aRoutes := make(map[string][]string)

	// Iterate through all routes defined in the service
	for _, route := range p.routes() {
		// Skip non-admin API routes
		if !strings.HasPrefix(route.Path, "/go-api/internal/admin") || strings.HasSuffix(route.Path, "ping") {
			continue
		}

		exist := false // Flag to check if route already has permission

		// Check if current route already has permission assigned
		for _, permission := range perms {
			if route.Method == permission.Method && route.Path == permission.Path {
				exist = true
				break // Exit loop if match found
			}
		}

		// If route already has permission, skip to next route
		if exist {
			continue
		}

		// If route has no permission, add it to available routes map
		// Key is HTTP method, value is slice of paths
		aRoutes[route.Method] = append(aRoutes[route.Method], route.Path)
	}

	// Return the map of available routes
	return aRoutes, nil
}

func (p permissionService) Create(ctx context.Context, permission *system.Permission) (errCode int, err error) {
	errCode = e.PermissionNameCantBeNull
	if permission.Name != "" {
		var perm *system.Permission

		errCode = e.PermissionNameExists
		perm, err = p.permRepo.Detail(ctx, &system.Permission{Name: permission.Name})
		if perm == nil {
			errCode = e.ERROR
			_, err = p.permRepo.Create(ctx, permission)
			if err == nil {
				errCode = e.SUCCESS
			}
		}
	}

	return
}

func (p permissionService) Delete(ctx context.Context, id uint) error {
	err := p.permRepo.DeleteByID(ctx, id)
	if err == nil {
		err = p.permRoleRepo.DeleteByPermID(ctx, id)
	}

	return err
}

func (p permissionService) Detail(ctx context.Context, id uint) (permission *system.Permission, errCode int, err error) {
	permission, err = p.permRepo.Detail(ctx, &system.Permission{Model: gorm.Model{ID: id}})
	if err != nil {
		return nil, e.PermissionNotFound, err
	}

	return permission, e.SUCCESS, nil
}

func (p permissionService) ListInGroup(ctx context.Context, permType string) (list map[string][]system.Permission, err error) {
	permissions, err := p.permRepo.ListByType(ctx, permType)
	if err != nil {
		return
	}

	list = make(map[string][]system.Permission)
	for _, perm := range permissions {
		list[perm.Group] = append(list[perm.Group], perm)
	}

	return
}

func (p permissionService) Update(ctx context.Context, permission *system.Permission) (errCode int, err error) {
	errCode = e.PermissionNotFound

	perm, err := p.permRepo.Detail(ctx, &system.Permission{Model: gorm.Model{ID: permission.ID}})
	if perm != nil {
		errCode = e.ERROR
		err = p.permRepo.Update(ctx, permission)
		if err == nil {
			errCode = e.SUCCESS
		}
	}

	return
}

// NewPermissionService creates a new PermissionService instance.
func NewPermissionService(redis *redis.Manager, logger *logger.Manager, db *gorm.DB, routes func() gin.RoutesInfo) PermissionService {
	return &permissionService{
		redis:        redis,
		logger:       logger,
		routes:       routes,
		permRepo:     repo.NewPermissionRepo(db, redis, logger),
		permRoleRepo: repo.NewPermissionRoleRepo(db, redis, logger),
	}
}
