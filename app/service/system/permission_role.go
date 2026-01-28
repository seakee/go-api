package system

import (
	"context"
	"github.com/seakee/go-api/app/pkg/e"
	"strings"
)

func (r roleService) UpdatePermission(ctx context.Context, roleId uint, permissionIds []uint) (int, error) {
	role, err := r.roleRepo.DetailByID(ctx, roleId)
	if role == nil {
		return e.RoleNotFound, err
	}

	roleName := strings.ToLower(role.Name)
	if roleName == "base" || roleName == "super_admin" {
		return e.RoleCanNotBeOperated, err
	}

	// Use map to deduplicate
	uniquePermIDs := make(map[uint]struct{}, len(permissionIds))
	for _, id := range permissionIds {
		uniquePermIDs[id] = struct{}{}
	}

	// Convert deduplicated IDs to slice
	permIDs := make([]uint, 0, len(uniquePermIDs))
	for id := range uniquePermIDs {
		permIDs = append(permIDs, id)
	}

	err = r.permissionRoleRepo.UpdateRolePermission(ctx, roleId, permIDs)
	if err != nil {
		return e.ERROR, err
	}

	return e.SUCCESS, nil
}

func (r roleService) PermissionList(ctx context.Context, roleId uint) (list []uint, err error) {
	return r.permissionRoleRepo.ListByRoleID(ctx, roleId)
}
