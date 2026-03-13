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
	"time"

	"github.com/seakee/go-api/app/model/system"
	repo "github.com/seakee/go-api/app/repository/system"
)

// mockUserRepo is a mock implementation of the UserRepo interface.
type mockUserRepo struct {
	DetailFunc             func(ctx context.Context, user *system.User) (*system.User, error)
	CreateFunc             func(ctx context.Context, user *system.User) (*system.User, error)
	DetailByEmailFunc      func(ctx context.Context, email string) (*system.User, error)
	DetailByPhoneFunc      func(ctx context.Context, phone string) (*system.User, error)
	DetailByIdentifierFunc func(ctx context.Context, identifier string) (*system.User, error)
	CreateOrUpdateFunc     func(ctx context.Context, user *system.User) (*system.User, error)
	DetailByIDFunc         func(ctx context.Context, id uint) (*system.User, error)
	ListByIDsFunc          func(ctx context.Context, ids []uint) ([]system.User, error)
	DeleteByIDFunc         func(ctx context.Context, id uint) error
	UpdateFunc             func(ctx context.Context, user *system.User) error
	UpdateIdentifierFunc   func(ctx context.Context, user *system.User) error
	UpdateTotpStatusFunc   func(ctx context.Context, user *system.User) error
	PaginateFunc           func(ctx context.Context, user *system.User, page, pageSize int) ([]system.User, error)
	GetOAuthUserFunc       func(ctx context.Context, oauthType, tenant, id string) (*system.User, error)
	BindOAuthIdentityFunc  func(ctx context.Context, input repo.OAuthBindInput) error
	CountFunc              func(ctx context.Context, user *system.User) (int64, error)
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

func (m *mockUserRepo) DetailByEmail(ctx context.Context, email string) (*system.User, error) {
	if m.DetailByEmailFunc != nil {
		return m.DetailByEmailFunc(ctx, email)
	}
	return nil, nil
}

func (m *mockUserRepo) DetailByPhone(ctx context.Context, phone string) (*system.User, error) {
	if m.DetailByPhoneFunc != nil {
		return m.DetailByPhoneFunc(ctx, phone)
	}
	return nil, nil
}

func (m *mockUserRepo) DetailByIdentifier(ctx context.Context, identifier string) (*system.User, error) {
	if m.DetailByIdentifierFunc != nil {
		return m.DetailByIdentifierFunc(ctx, identifier)
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

func (m *mockUserRepo) ListByIDs(ctx context.Context, ids []uint) ([]system.User, error) {
	if m.ListByIDsFunc != nil {
		return m.ListByIDsFunc(ctx, ids)
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

func (m *mockUserRepo) UpdateIdentifier(ctx context.Context, user *system.User) error {
	if m.UpdateIdentifierFunc != nil {
		return m.UpdateIdentifierFunc(ctx, user)
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

func (m *mockUserRepo) GetOAuthUser(ctx context.Context, oauthType, tenant, id string) (*system.User, error) {
	if m.GetOAuthUserFunc != nil {
		return m.GetOAuthUserFunc(ctx, oauthType, tenant, id)
	}
	return nil, nil
}

func (m *mockUserRepo) BindOAuthIdentity(ctx context.Context, input repo.OAuthBindInput) error {
	if m.BindOAuthIdentityFunc != nil {
		return m.BindOAuthIdentityFunc(ctx, input)
	}
	return nil
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
	ListRolesFunc     func(ctx context.Context, userID uint) (map[string]uint, error)
	HasPermissionFunc func(ctx context.Context, userID uint, permissionHash string) (bool, error)
}

func (m *mockAuthRepo) HasRole(ctx context.Context, userID uint, role string) (bool, error) {
	if m.HasRoleFunc != nil {
		return m.HasRoleFunc(ctx, userID, role)
	}
	return false, nil
}

func (m *mockAuthRepo) ListRoles(ctx context.Context, userID uint) (map[string]uint, error) {
	if m.ListRolesFunc != nil {
		return m.ListRolesFunc(ctx, userID)
	}
	return map[string]uint{}, nil
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

type mockUserIdentityRepo struct {
	DetailByProviderFunc    func(ctx context.Context, provider, tenant, subject string) (*system.UserIdentity, error)
	DetailByIDAndUserIDFunc func(ctx context.Context, id, userID uint) (*system.UserIdentity, error)
	ListByUserIDFunc        func(ctx context.Context, userID uint) ([]system.UserIdentity, error)
	CreateFunc              func(ctx context.Context, identity *system.UserIdentity) (*system.UserIdentity, error)
	UpdateLastLoginFunc     func(ctx context.Context, id uint, loginAt time.Time) error
	DeleteByIDAndUserIDFunc func(ctx context.Context, id, userID uint) error
	DeleteByUserIDFunc      func(ctx context.Context, userID uint) error
}

func (m *mockUserIdentityRepo) DetailByProvider(ctx context.Context, provider, tenant, subject string) (*system.UserIdentity, error) {
	if m.DetailByProviderFunc != nil {
		return m.DetailByProviderFunc(ctx, provider, tenant, subject)
	}
	return nil, nil
}

func (m *mockUserIdentityRepo) ListByUserID(ctx context.Context, userID uint) ([]system.UserIdentity, error) {
	if m.ListByUserIDFunc != nil {
		return m.ListByUserIDFunc(ctx, userID)
	}
	return nil, nil
}

func (m *mockUserIdentityRepo) DetailByIDAndUserID(ctx context.Context, id, userID uint) (*system.UserIdentity, error) {
	if m.DetailByIDAndUserIDFunc != nil {
		return m.DetailByIDAndUserIDFunc(ctx, id, userID)
	}
	return nil, nil
}

func (m *mockUserIdentityRepo) Create(ctx context.Context, identity *system.UserIdentity) (*system.UserIdentity, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, identity)
	}
	return identity, nil
}

func (m *mockUserIdentityRepo) UpdateLastLogin(ctx context.Context, id uint, loginAt time.Time) error {
	if m.UpdateLastLoginFunc != nil {
		return m.UpdateLastLoginFunc(ctx, id, loginAt)
	}
	return nil
}

func (m *mockUserIdentityRepo) DeleteByIDAndUserID(ctx context.Context, id, userID uint) error {
	if m.DeleteByIDAndUserIDFunc != nil {
		return m.DeleteByIDAndUserIDFunc(ctx, id, userID)
	}
	return nil
}

func (m *mockUserIdentityRepo) DeleteByUserID(ctx context.Context, userID uint) error {
	if m.DeleteByUserIDFunc != nil {
		return m.DeleteByUserIDFunc(ctx, userID)
	}
	return nil
}

type mockUserPasskeyRepo struct {
	DetailByCredentialIDFunc func(ctx context.Context, credentialID string) (*system.UserPasskey, error)
	DetailByIDAndUserIDFunc  func(ctx context.Context, id, userID uint) (*system.UserPasskey, error)
	ListByUserIDFunc         func(ctx context.Context, userID uint) ([]system.UserPasskey, error)
	CountByUserIDFunc        func(ctx context.Context, userID uint) (int64, error)
	CountByUserIDsFunc       func(ctx context.Context, userIDs []uint) (map[uint]int64, error)
	CreateFunc               func(ctx context.Context, passkey *system.UserPasskey) (*system.UserPasskey, error)
	UpdateSignCountFunc      func(ctx context.Context, id uint, signCount uint32) error
	UpdateLastUsedAtFunc     func(ctx context.Context, id uint, usedAt time.Time) error
	DeleteByIDAndUserIDFunc  func(ctx context.Context, id, userID uint) error
	DeleteByIDFunc           func(ctx context.Context, id uint) error
	DeleteByUserIDFunc       func(ctx context.Context, userID uint) error
}

func (m *mockUserPasskeyRepo) DetailByCredentialID(ctx context.Context, credentialID string) (*system.UserPasskey, error) {
	if m.DetailByCredentialIDFunc != nil {
		return m.DetailByCredentialIDFunc(ctx, credentialID)
	}
	return nil, nil
}

func (m *mockUserPasskeyRepo) DetailByIDAndUserID(ctx context.Context, id, userID uint) (*system.UserPasskey, error) {
	if m.DetailByIDAndUserIDFunc != nil {
		return m.DetailByIDAndUserIDFunc(ctx, id, userID)
	}
	return nil, nil
}

func (m *mockUserPasskeyRepo) ListByUserID(ctx context.Context, userID uint) ([]system.UserPasskey, error) {
	if m.ListByUserIDFunc != nil {
		return m.ListByUserIDFunc(ctx, userID)
	}
	return nil, nil
}

func (m *mockUserPasskeyRepo) CountByUserID(ctx context.Context, userID uint) (int64, error) {
	if m.CountByUserIDFunc != nil {
		return m.CountByUserIDFunc(ctx, userID)
	}
	return 0, nil
}

func (m *mockUserPasskeyRepo) CountByUserIDs(ctx context.Context, userIDs []uint) (map[uint]int64, error) {
	if m.CountByUserIDsFunc != nil {
		return m.CountByUserIDsFunc(ctx, userIDs)
	}
	return map[uint]int64{}, nil
}

func (m *mockUserPasskeyRepo) Create(ctx context.Context, passkey *system.UserPasskey) (*system.UserPasskey, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, passkey)
	}
	return passkey, nil
}

func (m *mockUserPasskeyRepo) UpdateSignCount(ctx context.Context, id uint, signCount uint32) error {
	if m.UpdateSignCountFunc != nil {
		return m.UpdateSignCountFunc(ctx, id, signCount)
	}
	return nil
}

func (m *mockUserPasskeyRepo) UpdateLastUsedAt(ctx context.Context, id uint, usedAt time.Time) error {
	if m.UpdateLastUsedAtFunc != nil {
		return m.UpdateLastUsedAtFunc(ctx, id, usedAt)
	}
	return nil
}

func (m *mockUserPasskeyRepo) DeleteByIDAndUserID(ctx context.Context, id, userID uint) error {
	if m.DeleteByIDAndUserIDFunc != nil {
		return m.DeleteByIDAndUserIDFunc(ctx, id, userID)
	}
	return nil
}

func (m *mockUserPasskeyRepo) DeleteByID(ctx context.Context, id uint) error {
	if m.DeleteByIDFunc != nil {
		return m.DeleteByIDFunc(ctx, id)
	}
	return nil
}

func (m *mockUserPasskeyRepo) DeleteByUserID(ctx context.Context, userID uint) error {
	if m.DeleteByUserIDFunc != nil {
		return m.DeleteByUserIDFunc(ctx, userID)
	}
	return nil
}

// mockOperationRecordRepo is a mock implementation of the OperationRecordRepo interface.
type mockOperationRecordRepo struct {
	PaginationFunc func(ctx context.Context, record *system.OperationRecord, page, pageSize int) ([]system.OperationRecord, int64, error)
	DetailFunc     func(ctx context.Context, id string) (*repo.OperationRecordDetail, error)
	CreateFunc     func(ctx context.Context, record *system.OperationRecord) error
}

func (m *mockOperationRecordRepo) Pagination(ctx context.Context, record *system.OperationRecord, page, pageSize int) ([]system.OperationRecord, int64, error) {
	if m.PaginationFunc != nil {
		return m.PaginationFunc(ctx, record, page, pageSize)
	}
	return nil, 0, nil
}

func (m *mockOperationRecordRepo) Detail(ctx context.Context, id string) (*repo.OperationRecordDetail, error) {
	if m.DetailFunc != nil {
		return m.DetailFunc(ctx, id)
	}
	return nil, nil
}

func (m *mockOperationRecordRepo) Create(ctx context.Context, record *system.OperationRecord) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, record)
	}
	return nil
}
