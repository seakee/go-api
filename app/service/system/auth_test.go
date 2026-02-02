package system

import (
	"context"
	"errors"
	"testing"

	"gorm.io/gorm"

	"github.com/seakee/go-api/app/model/system"
	"github.com/seakee/go-api/app/pkg/e"
)

// TestAuthService_Profile tests fetching user profile.
func TestAuthService_Profile(t *testing.T) {
	tests := []struct {
		name           string
		userID         uint
		mockUser       *system.User
		mockErr        error
		wantErrCode    int
		wantErr        bool
		wantUserFields []string
	}{
		{
			name:   "get user profile successfully",
			userID: 1,
			mockUser: &system.User{
				Model:    gorm.Model{ID: 1},
				UserName: "testuser",
				Avatar:   "https://example.com/avatar.png",
			},
			mockErr:        nil,
			wantErrCode:    0,
			wantErr:        false,
			wantUserFields: []string{"id", "user_name", "avatar"},
		},
		{
			name:           "user not found",
			userID:         999,
			mockUser:       nil,
			mockErr:        nil,
			wantErrCode:    e.AccountNotFound,
			wantErr:        false,
			wantUserFields: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserRepo := &mockUserRepo{
				DetailByIDFunc: func(ctx context.Context, id uint) (*system.User, error) {
					if id == tt.userID {
						return tt.mockUser, tt.mockErr
					}
					return nil, nil
				},
			}

			svc := &authService{
				userRepo: mockUserRepo,
			}

			user, errCode, err := svc.Profile(context.Background(), tt.userID)

			if (err != nil) != tt.wantErr {
				t.Errorf("Profile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if errCode != tt.wantErrCode {
				t.Errorf("Profile() errCode = %v, want %v", errCode, tt.wantErrCode)
				return
			}
			if tt.wantUserFields != nil {
				for _, field := range tt.wantUserFields {
					if _, ok := user[field]; !ok {
						t.Errorf("Profile() missing field %s in user map", field)
					}
				}
			}
		})
	}
}

// TestAuthService_HasPermission tests permission checks.
func TestAuthService_HasPermission(t *testing.T) {
	tests := []struct {
		name           string
		userID         uint
		permissionHash string
		mockResult     bool
		mockErr        error
		want           bool
		wantErr        bool
	}{
		{
			name:           "user has permission",
			userID:         1,
			permissionHash: "GET:/api/v1/users",
			mockResult:     true,
			mockErr:        nil,
			want:           true,
			wantErr:        false,
		},
		{
			name:           "user lacks permission",
			userID:         1,
			permissionHash: "DELETE:/api/v1/users",
			mockResult:     false,
			mockErr:        nil,
			want:           false,
			wantErr:        false,
		},
		{
			name:           "error when checking permission",
			userID:         1,
			permissionHash: "GET:/api/v1/users",
			mockResult:     false,
			mockErr:        errors.New("database error"),
			want:           false,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAuthRepo := &mockAuthRepo{
				HasPermissionFunc: func(ctx context.Context, userID uint, permissionHash string) (bool, error) {
					return tt.mockResult, tt.mockErr
				},
			}

			svc := &authService{
				authRepo: mockAuthRepo,
			}

			got, err := svc.HasPermission(context.Background(), tt.userID, tt.permissionHash)

			if (err != nil) != tt.wantErr {
				t.Errorf("HasPermission() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("HasPermission() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestAuthService_HasRole tests role checks.
func TestAuthService_HasRole(t *testing.T) {
	tests := []struct {
		name       string
		userID     uint
		role       string
		mockResult bool
		mockErr    error
		want       bool
		wantErr    bool
	}{
		{
			name:       "user has role",
			userID:     1,
			role:       "admin",
			mockResult: true,
			mockErr:    nil,
			want:       true,
			wantErr:    false,
		},
		{
			name:       "user lacks role",
			userID:     1,
			role:       "super_admin",
			mockResult: false,
			mockErr:    nil,
			want:       false,
			wantErr:    false,
		},
		{
			name:       "error when checking role",
			userID:     1,
			role:       "admin",
			mockResult: false,
			mockErr:    errors.New("database error"),
			want:       false,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAuthRepo := &mockAuthRepo{
				HasRoleFunc: func(ctx context.Context, userID uint, role string) (bool, error) {
					return tt.mockResult, tt.mockErr
				},
			}

			svc := &authService{
				authRepo: mockAuthRepo,
			}

			got, err := svc.HasRole(context.Background(), tt.userID, tt.role)

			if (err != nil) != tt.wantErr {
				t.Errorf("HasRole() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("HasRole() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestAuthService_TfaStatus tests two-factor status lookup.
func TestAuthService_TfaStatus(t *testing.T) {
	tests := []struct {
		name        string
		userID      uint
		mockUser    *system.User
		mockErr     error
		wantEnable  bool
		wantErrCode int
		wantErr     bool
	}{
		{
			name:   "two-factor enabled",
			userID: 1,
			mockUser: &system.User{
				Model:       gorm.Model{ID: 1},
				TotpEnabled: true,
			},
			mockErr:     nil,
			wantEnable:  true,
			wantErrCode: 0,
			wantErr:     false,
		},
		{
			name:   "two-factor disabled",
			userID: 1,
			mockUser: &system.User{
				Model:       gorm.Model{ID: 1},
				TotpEnabled: false,
			},
			mockErr:     nil,
			wantEnable:  false,
			wantErrCode: 0,
			wantErr:     false,
		},
		{
			name:        "user not found",
			userID:      999,
			mockUser:    nil,
			mockErr:     nil,
			wantEnable:  false,
			wantErrCode: e.AccountNotFound,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserRepo := &mockUserRepo{
				DetailFunc: func(ctx context.Context, user *system.User) (*system.User, error) {
					if user.ID == tt.userID {
						return tt.mockUser, tt.mockErr
					}
					return nil, nil
				},
			}

			svc := &authService{
				userRepo: mockUserRepo,
			}

			enable, errCode, err := svc.TfaStatus(context.Background(), tt.userID)

			if (err != nil) != tt.wantErr {
				t.Errorf("TfaStatus() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if errCode != tt.wantErrCode {
				t.Errorf("TfaStatus() errCode = %v, want %v", errCode, tt.wantErrCode)
				return
			}
			if enable != tt.wantEnable {
				t.Errorf("TfaStatus() enable = %v, want %v", enable, tt.wantEnable)
			}
		})
	}
}

// TestAuthService_UserMenuList tests fetching the user menu list.
func TestAuthService_UserMenuList(t *testing.T) {
	tests := []struct {
		name           string
		userID         uint
		isSuperAdmin   bool
		mockMenuList   system.MenuList
		mockRoleErr    error
		mockMenuErr    error
		wantMenuCount  int
		wantErr        bool
	}{
		{
			name:         "regular user gets menu",
			userID:       1,
			isSuperAdmin: false,
			mockMenuList: system.MenuList{
				{Model: gorm.Model{ID: 1}, Name: "Dashboard"},
				{Model: gorm.Model{ID: 2}, Name: "Users"},
			},
			mockRoleErr:   nil,
			mockMenuErr:   nil,
			wantMenuCount: 2,
			wantErr:       false,
		},
		{
			name:         "super admin gets menu",
			userID:       1,
			isSuperAdmin: true,
			mockMenuList: system.MenuList{
				{Model: gorm.Model{ID: 1}, Name: "Dashboard"},
				{Model: gorm.Model{ID: 2}, Name: "Users"},
				{Model: gorm.Model{ID: 3}, Name: "Settings"},
			},
			mockRoleErr:   nil,
			mockMenuErr:   nil,
			wantMenuCount: 3,
			wantErr:       false,
		},
		{
			name:          "failed to get menu",
			userID:        1,
			isSuperAdmin:  false,
			mockMenuList:  nil,
			mockRoleErr:   nil,
			mockMenuErr:   errors.New("database error"),
			wantMenuCount: 0,
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAuthRepo := &mockAuthRepo{
				HasRoleFunc: func(ctx context.Context, userID uint, role string) (bool, error) {
					if role == "super_admin" {
						return tt.isSuperAdmin, tt.mockRoleErr
					}
					return false, nil
				},
			}

			mockMenuRepo := &mockMenuRepo{
				UserMenuListFunc: func(ctx context.Context, userID uint, isSuperAdmin bool) (system.MenuList, error) {
					return tt.mockMenuList, tt.mockMenuErr
				},
			}

			svc := &authService{
				authRepo: mockAuthRepo,
				menuRepo: mockMenuRepo,
			}

			list, err := svc.UserMenuList(context.Background(), tt.userID)

			if (err != nil) != tt.wantErr {
				t.Errorf("UserMenuList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(list) != tt.wantMenuCount {
				t.Errorf("UserMenuList() got %d menus, want %d", len(list), tt.wantMenuCount)
			}
		})
	}
}
