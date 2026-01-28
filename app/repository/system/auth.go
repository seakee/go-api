package system

import (
	"context"
	"github.com/pkg/errors"

	redigo "github.com/gomodule/redigo/redis"
	"github.com/seakee/go-api/app/model/system"
	"github.com/seakee/go-api/app/model/system/permission"
	"github.com/seakee/go-api/app/model/system/role"
	"github.com/sk-pkg/logger"
	"github.com/sk-pkg/redis"
	"github.com/sk-pkg/util"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const (
	roleCacheSuffix = ":roles"
	permCacheSuffix = ":permissions"
	defaultCacheTTL = 7200 // 2 hours
)

// AuthRepo defines the authentication repository interface.
type AuthRepo interface {
	HasRole(ctx context.Context, userID uint, role string) (bool, error)
	HasPermission(ctx context.Context, userID uint, permissionHash string) (bool, error)
}

// authRepo implements the AuthRepo interface.
type authRepo struct {
	db     *gorm.DB
	redis  *redis.Manager
	logger *logger.Manager
}

// HasRole checks if the user has the specified role.
func (a authRepo) HasRole(ctx context.Context, userID uint, role string) (bool, error) {
	// Construct role cache key
	roleCacheKey := util.SpliceStr(userCachePrefix, roleCacheSuffix)

	// Get user role mapping
	userRoleList, err := a.getUserRoles(ctx, roleCacheKey)
	if err != nil {
		return false, err
	}

	// Check if user exists in role mapping
	roles, exists := userRoleList[userID]
	if !exists {
		return false, nil
	}

	// Check if user has the specified role
	_, hasRole := roles[role]
	return hasRole, nil
}

// HasPermission checks if the user has the specified permission.
func (a authRepo) HasPermission(ctx context.Context, userID uint, permissionHash string) (bool, error) {
	// Construct permission cache key
	permCacheKey := util.SpliceStr(userCachePrefix, permCacheSuffix)

	// Get user permission mapping
	userPermissionList, err := a.getUserPermissions(ctx, permCacheKey)
	if err != nil {
		return false, err
	}

	// Check if user exists in permission mapping
	permissions, exists := userPermissionList[userID]
	if !exists {
		return false, nil
	}

	// Check if user has the specified permission
	_, hasPermission := permissions[permissionHash]
	return hasPermission, nil
}

// getUserRoles retrieves the user role mapping.
func (a authRepo) getUserRoles(ctx context.Context, cacheKey string) (map[uint]map[string]uint, error) {
	var userRoleList map[uint]map[string]uint

	// Try to get user role mapping from Redis cache
	err := a.redis.GetJSON(cacheKey, &userRoleList)
	if err != nil && !errors.Is(err, redigo.ErrNil) {
		a.logger.Error(ctx, "Failed to get user roles from Redis", zap.Error(err))
	}

	// If cache is empty, rebuild the user role mapping
	if errors.Is(err, redigo.ErrNil) {
		userRoleList = a.userRoleMap(ctx)

		// Store the newly built user role mapping in Redis cache
		if err = a.redis.SetJSON(cacheKey, userRoleList, defaultCacheTTL); err != nil {
			a.logger.Error(ctx, "Failed to set user roles in Redis", zap.Error(err))
		}
	}

	return userRoleList, nil
}

// getUserPermissions retrieves the user permission mapping.
func (a authRepo) getUserPermissions(ctx context.Context, cacheKey string) (map[uint]map[string]uint, error) {
	var userPermissionList map[uint]map[string]uint

	// Try to get user permission mapping from Redis cache
	err := a.redis.GetJSON(cacheKey, &userPermissionList)
	if err != nil && !errors.Is(err, redigo.ErrNil) {
		a.logger.Error(ctx, "Failed to get user permissions from Redis", zap.Error(err))
	}

	// If cache is empty, rebuild the user permission mapping
	if errors.Is(err, redigo.ErrNil) {
		userPermissionList = a.userPermissionMap(ctx)

		// Store the newly built user permission mapping in Redis cache
		if err = a.redis.SetJSON(cacheKey, userPermissionList, defaultCacheTTL); err != nil {
			a.logger.Error(ctx, "Failed to set user permissions in Redis", zap.Error(err))
		}
	}

	return userPermissionList, nil
}

// userRoleMap builds the user role mapping.
func (a authRepo) userRoleMap(ctx context.Context) map[uint]map[string]uint {
	// Get role name mapping
	roleNameMap := a.getRoleNameMap(ctx)
	userRoles := make(map[uint]map[string]uint)

	ru := &role.User{}
	// Get all user role associations
	list, err := ru.List(ctx, a.db)
	if err != nil {
		a.logger.Error(ctx, "Failed to get role user list", zap.Error(err))
		return userRoles
	}

	// Iterate through user role associations to build user role mapping
	for _, roleUser := range list {
		if _, exists := userRoles[roleUser.UserId]; !exists {
			userRoles[roleUser.UserId] = make(map[string]uint)
		}

		if roleName, exists := roleNameMap[roleUser.RoleId]; exists {
			userRoles[roleUser.UserId][roleName] = roleUser.RoleId
		}
	}

	return userRoles
}

// getRoleNameMap retrieves the role name mapping.
func (a authRepo) getRoleNameMap(ctx context.Context) map[uint]string {
	roleNameMap := make(map[uint]string)
	r := &system.Role{}
	// Get all roles
	roleList, err := r.List(ctx, a.db)
	if err != nil {
		a.logger.Error(ctx, "Failed to get role list", zap.Error(err))
		return roleNameMap
	}

	// Build role ID to role name mapping
	for _, item := range roleList {
		roleNameMap[item.ID] = item.Name
	}

	return roleNameMap
}

// userPermissionMap builds the user permission mapping.
func (a authRepo) userPermissionMap(ctx context.Context) map[uint]map[string]uint {
	userPermissions := make(map[uint]map[string]uint)
	// Get permission hash mapping
	permHashMap := a.getPermissionHashMap(ctx)
	// Get role permission mapping
	permList := a.getRolePermissionMap(ctx)

	ru := &role.User{}
	// Get all user role associations
	list, err := ru.List(ctx, a.db)
	if err != nil {
		a.logger.Error(ctx, "Failed to get role user list", zap.Error(err))
		return userPermissions
	}

	// Iterate through user role associations to build user permission mapping
	for _, roleUser := range list {
		if _, exists := userPermissions[roleUser.UserId]; !exists {
			userPermissions[roleUser.UserId] = make(map[string]uint)
		}

		if perms, exists := permList[roleUser.RoleId]; exists {
			for _, permID := range perms {
				if hash, exists := permHashMap[permID]; exists {
					userPermissions[roleUser.UserId][hash] = permID
				}
			}
		}
	}

	return userPermissions
}

// getPermissionHashMap retrieves the permission hash mapping.
func (a authRepo) getPermissionHashMap(ctx context.Context) map[uint]string {
	permHashMap := make(map[uint]string)
	p := &system.Permission{}
	// Get all permissions
	permissions, err := p.List(ctx, a.db)
	if err != nil {
		a.logger.Error(ctx, "Failed to get permission list", zap.Error(err))
		return permHashMap
	}

	// Build permission ID to permission hash mapping
	for _, perm := range permissions {
		permHashMap[perm.ID] = util.MD5(perm.Method + perm.Path)
	}

	return permHashMap
}

// getRolePermissionMap retrieves the role permission mapping.
func (a authRepo) getRolePermissionMap(ctx context.Context) map[uint][]uint {
	permList := make(map[uint][]uint)
	pr := &permission.Role{}
	// Get all role permission associations
	permRoles, err := pr.List(ctx, a.db)
	if err != nil {
		a.logger.Error(ctx, "Failed to get permission role list", zap.Error(err))
		return permList
	}

	// Build role ID to permission ID list mapping
	for _, permRole := range permRoles {
		permList[permRole.RoleId] = append(permList[permRole.RoleId], permRole.PermissionId)
	}

	return permList
}

// NewAuthRepo creates a new AuthRepo instance.
func NewAuthRepo(db *gorm.DB, redis *redis.Manager, logger *logger.Manager) AuthRepo {
	return &authRepo{db: db, redis: redis, logger: logger}
}
