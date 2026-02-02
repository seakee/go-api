// Package system provides mock implementations for the system module, used for unit tests.
//
// Note: The following interfaces include an unexported i() method to prevent implementations in external packages:
// - RoleRepo
// - RoleUserRepo
// - PermissionRepo
// - PermissionRoleRepo
//
// Mocks for these interfaces should be provided in the repository/system package or covered by integration tests.
// This file only provides interfaces that can be mocked within the service package.
package system

import (
	"context"

	"github.com/seakee/go-api/app/model/system"
)

// mockUserRepo is a mock implementation of the UserRepo interface.
type mockUserRepo struct {
	DetailFunc           func(ctx context.Context, user *system.User) (*system.User, error)
	CreateFunc           func(ctx context.Context, user *system.User) (*system.User, error)
	DetailByAccountFunc  func(ctx context.Context, account string) (*system.User, error)
	CreateOrUpdateFunc   func(ctx context.Context, user *system.User) (*system.User, error)
	DetailByIDFunc       func(ctx context.Context, id uint) (*system.User, error)
	DeleteByIDFunc       func(ctx context.Context, id uint) error
	UpdateFunc           func(ctx context.Context, user *system.User) error
	UpdateTotpStatusFunc func(ctx context.Context, user *system.User) error
	PaginateFunc         func(ctx context.Context, user *system.User, page, pageSize int) ([]system.User, error)
	GetOAuthUserFunc     func(ctx context.Context, oauthType, id string) (*system.User, error)
	CountFunc            func(ctx context.Context, user *system.User) (int64, error)
}

func (m *mockUserRepo) Detail(ctx context.Context, user *system.User) (*system.User, error) {
	if m.DetailFunc != nil {
		return m.DetailFunc(ctx, user)
	}
	return nil, nil
}

func (m *mockUserRepo) Create(ctx context.Context, user *system.User) (*system.User, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, user)
	}
	return user, nil
}

func (m *mockUserRepo) DetailByAccount(ctx context.Context, account string) (*system.User, error) {
	if m.DetailByAccountFunc != nil {
		return m.DetailByAccountFunc(ctx, account)
	}
	return nil, nil
}

func (m *mockUserRepo) CreateOrUpdate(ctx context.Context, user *system.User) (*system.User, error) {
	if m.CreateOrUpdateFunc != nil {
		return m.CreateOrUpdateFunc(ctx, user)
	}
	return user, nil
}

func (m *mockUserRepo) DetailByID(ctx context.Context, id uint) (*system.User, error) {
	if m.DetailByIDFunc != nil {
		return m.DetailByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *mockUserRepo) DeleteByID(ctx context.Context, id uint) error {
	if m.DeleteByIDFunc != nil {
		return m.DeleteByIDFunc(ctx, id)
	}
	return nil
}

func (m *mockUserRepo) Update(ctx context.Context, user *system.User) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, user)
	}
	return nil
}

func (m *mockUserRepo) UpdateTotpStatus(ctx context.Context, user *system.User) error {
	if m.UpdateTotpStatusFunc != nil {
		return m.UpdateTotpStatusFunc(ctx, user)
	}
	return nil
}

func (m *mockUserRepo) Paginate(ctx context.Context, user *system.User, page, pageSize int) ([]system.User, error) {
	if m.PaginateFunc != nil {
		return m.PaginateFunc(ctx, user, page, pageSize)
	}
	return nil, nil
}

func (m *mockUserRepo) GetOAuthUser(ctx context.Context, oauthType, id string) (*system.User, error) {
	if m.GetOAuthUserFunc != nil {
		return m.GetOAuthUserFunc(ctx, oauthType, id)
	}
	return nil, nil
}

func (m *mockUserRepo) Count(ctx context.Context, user *system.User) (int64, error) {
	if m.CountFunc != nil {
		return m.CountFunc(ctx, user)
	}
	return 0, nil
}

// mockAuthRepo is a mock implementation of the AuthRepo interface.
type mockAuthRepo struct {
	HasRoleFunc       func(ctx context.Context, userID uint, role string) (bool, error)
	HasPermissionFunc func(ctx context.Context, userID uint, permissionHash string) (bool, error)
}

func (m *mockAuthRepo) HasRole(ctx context.Context, userID uint, role string) (bool, error) {
	if m.HasRoleFunc != nil {
		return m.HasRoleFunc(ctx, userID, role)
	}
	return false, nil
}

func (m *mockAuthRepo) HasPermission(ctx context.Context, userID uint, permissionHash string) (bool, error) {
	if m.HasPermissionFunc != nil {
		return m.HasPermissionFunc(ctx, userID, permissionHash)
	}
	return false, nil
}

// mockMenuRepo is a mock implementation of the MenuRepo interface.
type mockMenuRepo struct {
	UserMenuListFunc func(ctx context.Context, userID uint, isSuperAdmin bool) (system.MenuList, error)
	ListFunc         func(ctx context.Context) (system.MenuList, error)
	DetailByIDFunc   func(ctx context.Context, id uint) (*system.Menu, error)
	DetailFunc       func(ctx context.Context, menu *system.Menu) (*system.Menu, error)
	CreateFunc       func(ctx context.Context, menu *system.Menu) (uint, error)
	DeleteByIDFunc   func(ctx context.Context, id uint) error
	UpdateFunc       func(ctx context.Context, menu *system.Menu) error
	CountFunc        func(ctx context.Context, menu *system.Menu) (int64, error)
}

func (m *mockMenuRepo) UserMenuList(ctx context.Context, userID uint, isSuperAdmin bool) (system.MenuList, error) {
	if m.UserMenuListFunc != nil {
		return m.UserMenuListFunc(ctx, userID, isSuperAdmin)
	}
	return nil, nil
}

func (m *mockMenuRepo) List(ctx context.Context) (system.MenuList, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx)
	}
	return nil, nil
}

func (m *mockMenuRepo) DetailByID(ctx context.Context, id uint) (*system.Menu, error) {
	if m.DetailByIDFunc != nil {
		return m.DetailByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *mockMenuRepo) Detail(ctx context.Context, menu *system.Menu) (*system.Menu, error) {
	if m.DetailFunc != nil {
		return m.DetailFunc(ctx, menu)
	}
	return nil, nil
}

func (m *mockMenuRepo) Create(ctx context.Context, menu *system.Menu) (uint, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, menu)
	}
	return 1, nil
}

func (m *mockMenuRepo) DeleteByID(ctx context.Context, id uint) error {
	if m.DeleteByIDFunc != nil {
		return m.DeleteByIDFunc(ctx, id)
	}
	return nil
}

func (m *mockMenuRepo) Update(ctx context.Context, menu *system.Menu) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, menu)
	}
	return nil
}

func (m *mockMenuRepo) Count(ctx context.Context, menu *system.Menu) (int64, error) {
	if m.CountFunc != nil {
		return m.CountFunc(ctx, menu)
	}
	return 0, nil
}

// mockOperationRecordRepo is a mock implementation of the OperationRecordRepo interface.
type mockOperationRecordRepo struct {
	PaginationFunc  func(ctx context.Context, record *system.OperationRecord, page, pageSize int) ([]system.OperationRecord, int64, error)
	InteractionFunc func(ctx context.Context, id string) (any, error)
	CreateFunc      func(ctx context.Context, record *system.OperationRecord) error
}

func (m *mockOperationRecordRepo) Pagination(ctx context.Context, record *system.OperationRecord, page, pageSize int) ([]system.OperationRecord, int64, error) {
	if m.PaginationFunc != nil {
		return m.PaginationFunc(ctx, record, page, pageSize)
	}
	return nil, 0, nil
}

func (m *mockOperationRecordRepo) Interaction(ctx context.Context, id string) (any, error) {
	if m.InteractionFunc != nil {
		return m.InteractionFunc(ctx, id)
	}
	return nil, nil
}

func (m *mockOperationRecordRepo) Create(ctx context.Context, record *system.OperationRecord) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, record)
	}
	return nil
}
