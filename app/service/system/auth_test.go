package system

import (
	"context"
	"errors"
	"testing"
	"time"

	"gorm.io/gorm"

	"github.com/seakee/go-api/app/model/system"
	"github.com/seakee/go-api/app/pkg/e"
	"github.com/seakee/go-api/app/pkg/totp"
	"github.com/sk-pkg/util"
)

// TestAuthService_Profile tests fetching user profile.
func TestAuthService_Profile(t *testing.T) {
	tests := []struct {
		name           string
		userID         uint
		mockUser       *system.User
		mockRoles      map[string]uint
		mockErr        error
		mockRolesErr   error
		wantErrCode    int
		wantErr        bool
		wantUserFields []string
		wantRoleName   string
	}{
		{
			name:   "get user profile successfully",
			userID: 1,
			mockUser: &system.User{
				Model:    gorm.Model{ID: 1},
				UserName: "testuser",
				Avatar:   "https://example.com/avatar.png",
			},
			mockRoles:      map[string]uint{"base": 1, "editor": 2},
			mockErr:        nil,
			mockRolesErr:   nil,
			wantErrCode:    0,
			wantErr:        false,
			wantUserFields: []string{"id", "user_name", "avatar", "role_name"},
			wantRoleName:   "管理员",
		},
		{
			name:   "super admin profile",
			userID: 2,
			mockUser: &system.User{
				Model:    gorm.Model{ID: 2},
				UserName: "super",
				Avatar:   "https://example.com/super.png",
			},
			mockRoles:      map[string]uint{"base": 1, "super_admin": 3},
			mockErr:        nil,
			mockRolesErr:   nil,
			wantErrCode:    0,
			wantErr:        false,
			wantUserFields: []string{"id", "user_name", "avatar", "role_name"},
			wantRoleName:   "超级管理员",
		},
		{
			name:   "base only profile",
			userID: 3,
			mockUser: &system.User{
				Model:    gorm.Model{ID: 3},
				UserName: "base_user",
				Avatar:   "https://example.com/base.png",
			},
			mockRoles:      map[string]uint{"base": 1},
			mockErr:        nil,
			mockRolesErr:   nil,
			wantErrCode:    0,
			wantErr:        false,
			wantUserFields: []string{"id", "user_name", "avatar", "role_name"},
			wantRoleName:   "普通用户",
		},
		{
			name:           "list roles failed",
			userID:         4,
			mockUser:       &system.User{Model: gorm.Model{ID: 4}, UserName: "err_user"},
			mockRoles:      nil,
			mockErr:        nil,
			mockRolesErr:   errors.New("list roles error"),
			wantErrCode:    e.ERROR,
			wantErr:        true,
			wantUserFields: nil,
		},
		{
			name:           "user not found",
			userID:         999,
			mockUser:       nil,
			mockRoles:      nil,
			mockErr:        nil,
			mockRolesErr:   nil,
			wantErrCode:    e.UserNotFound,
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
				authRepo: &mockAuthRepo{
					ListRolesFunc: func(ctx context.Context, userID uint) (map[string]uint, error) {
						if userID == tt.userID {
							return tt.mockRoles, tt.mockRolesErr
						}
						return map[string]uint{}, nil
					},
				},
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

				if tt.wantRoleName != "" {
					if got, ok := user["role_name"]; !ok || got != tt.wantRoleName {
						t.Errorf("Profile() role_name = %v, want %v", got, tt.wantRoleName)
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
			wantErrCode: e.UserNotFound,
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
		name          string
		userID        uint
		isSuperAdmin  bool
		mockMenuList  system.MenuList
		mockRoleErr   error
		mockMenuErr   error
		wantMenuCount int
		wantErr       bool
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

// TestAuthService_UpdatePassword tests update password with and without TFA.
func TestAuthService_UpdatePassword(t *testing.T) {
	generator := totp.NewGenerator("go-api-admin")
	totpKey := "JBSWY3DPEHPK3PXP"
	validTotpCode := generator.GenerateTOTPCode(totpKey, time.Now())

	tests := []struct {
		name        string
		totpCode    string
		oldPassword string
		password    string
		mockUser    *system.User
		wantErrCode int
		wantErr     bool
		wantUpdated bool
	}{
		{
			name:        "update password successfully when tfa disabled",
			oldPassword: "old-md5-password",
			password:    "new-md5-password",
			mockUser: &system.User{
				Model:       gorm.Model{ID: 1},
				TotpEnabled: false,
				Salt:        "salt",
				Password:    util.MD5("old-md5-password" + "salt"),
			},
			wantErrCode: 0,
			wantErr:     false,
			wantUpdated: true,
		},
		{
			name:        "old password mismatch when tfa disabled",
			oldPassword: "wrong-password",
			password:    "new-md5-password",
			mockUser: &system.User{
				Model:       gorm.Model{ID: 1},
				TotpEnabled: false,
				Salt:        "salt",
				Password:    util.MD5("old-md5-password" + "salt"),
			},
			wantErrCode: e.IdentifierOrPasswordFail,
			wantErr:     false,
			wantUpdated: false,
		},
		{
			name:     "totp code required when tfa enabled",
			password: "new-md5-password",
			mockUser: &system.User{
				Model:       gorm.Model{ID: 1},
				TotpEnabled: true,
				TotpKey:     totpKey,
			},
			wantErrCode: e.TotpCodeCanNotBeNull,
			wantErr:     false,
			wantUpdated: false,
		},
		{
			name:     "update password successfully when tfa enabled",
			totpCode: validTotpCode,
			password: "new-md5-password",
			mockUser: &system.User{
				Model:       gorm.Model{ID: 1},
				TotpEnabled: true,
				TotpKey:     totpKey,
			},
			wantErrCode: 0,
			wantErr:     false,
			wantUpdated: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updated := false
			mockUserRepo := &mockUserRepo{
				DetailFunc: func(ctx context.Context, user *system.User) (*system.User, error) {
					if user.ID == 1 {
						return tt.mockUser, nil
					}
					return nil, nil
				},
				UpdateFunc: func(ctx context.Context, user *system.User) error {
					updated = true
					if user.Password != tt.password {
						t.Errorf("Update() password = %s, want %s", user.Password, tt.password)
					}
					return nil
				},
			}

			svc := &authService{
				userRepo: mockUserRepo,
				totp:     generator,
			}

			errCode, err := svc.UpdatePassword(context.Background(), 1, tt.totpCode, tt.oldPassword, tt.password)

			if (err != nil) != tt.wantErr {
				t.Errorf("UpdatePassword() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if errCode != tt.wantErrCode {
				t.Errorf("UpdatePassword() errCode = %v, want %v", errCode, tt.wantErrCode)
			}

			if updated != tt.wantUpdated {
				t.Errorf("UpdatePassword() updated = %v, want %v", updated, tt.wantUpdated)
			}
		})
	}
}

// TestAuthService_UpdateIdentifier tests update identifier with and without TFA.
func TestAuthService_UpdateIdentifier(t *testing.T) {
	generator := totp.NewGenerator("go-api-admin")
	totpKey := "JBSWY3DPEHPK3PXP"
	validTotpCode := generator.GenerateTOTPCode(totpKey, time.Now())

	tests := []struct {
		name          string
		totpCode      string
		password      string
		email         string
		phone         string
		mockUser      *system.User
		existingEmail *system.User
		existingPhone *system.User
		wantErrCode   int
		wantErr       bool
		wantUpdated   bool
	}{
		{
			name:     "update identifier successfully when tfa disabled",
			password: "current-md5-password",
			email:    "new@example.com",
			mockUser: &system.User{
				Model:       gorm.Model{ID: 1},
				TotpEnabled: false,
				Salt:        "salt",
				Password:    util.MD5("current-md5-password" + "salt"),
			},
			wantErrCode: 0,
			wantErr:     false,
			wantUpdated: true,
		},
		{
			name:     "password mismatch when tfa disabled",
			password: "wrong-password",
			email:    "new@example.com",
			mockUser: &system.User{
				Model:       gorm.Model{ID: 1},
				TotpEnabled: false,
				Salt:        "salt",
				Password:    util.MD5("current-md5-password" + "salt"),
			},
			wantErrCode: e.IdentifierOrPasswordFail,
			wantErr:     false,
			wantUpdated: false,
		},
		{
			name:  "totp code required when tfa enabled",
			email: "new@example.com",
			mockUser: &system.User{
				Model:       gorm.Model{ID: 1},
				TotpEnabled: true,
				TotpKey:     totpKey,
			},
			wantErrCode: e.TotpCodeCanNotBeNull,
			wantErr:     false,
			wantUpdated: false,
		},
		{
			name:     "update identifier successfully when tfa enabled",
			totpCode: validTotpCode,
			email:    "new@example.com",
			mockUser: &system.User{
				Model:       gorm.Model{ID: 1},
				TotpEnabled: true,
				TotpKey:     totpKey,
			},
			wantErrCode: 0,
			wantErr:     false,
			wantUpdated: true,
		},
		{
			name:     "email exists",
			password: "current-md5-password",
			email:    "duplicated@example.com",
			mockUser: &system.User{
				Model:       gorm.Model{ID: 1},
				TotpEnabled: false,
				Salt:        "salt",
				Password:    util.MD5("current-md5-password" + "salt"),
			},
			existingEmail: &system.User{Model: gorm.Model{ID: 2}, Email: "duplicated@example.com"},
			wantErrCode:   e.IdentifierExists,
			wantErr:       false,
			wantUpdated:   false,
		},
		{
			name:     "phone exists",
			password: "current-md5-password",
			phone:    "+8613800000001",
			mockUser: &system.User{
				Model:       gorm.Model{ID: 1},
				TotpEnabled: false,
				Salt:        "salt",
				Password:    util.MD5("current-md5-password" + "salt"),
			},
			existingPhone: &system.User{Model: gorm.Model{ID: 2}, Phone: "+8613800000001"},
			wantErrCode:   e.IdentifierExists,
			wantErr:       false,
			wantUpdated:   false,
		},
		{
			name:     "switch from email to phone successfully",
			password: "current-md5-password",
			email:    "",
			phone:    "+8613800000002",
			mockUser: &system.User{
				Model:       gorm.Model{ID: 1},
				TotpEnabled: false,
				Salt:        "salt",
				Password:    util.MD5("current-md5-password" + "salt"),
				Email:       "old@example.com",
			},
			wantErrCode: e.SUCCESS,
			wantErr:     false,
			wantUpdated: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updated := false
			mockUserRepo := &mockUserRepo{
				DetailFunc: func(ctx context.Context, user *system.User) (*system.User, error) {
					if user.ID == 1 {
						return tt.mockUser, nil
					}
					return nil, nil
				},
				DetailByEmailFunc: func(ctx context.Context, email string) (*system.User, error) {
					if email == tt.email {
						return tt.existingEmail, nil
					}
					return nil, nil
				},
				DetailByPhoneFunc: func(ctx context.Context, phone string) (*system.User, error) {
					if phone == tt.phone {
						return tt.existingPhone, nil
					}
					return nil, nil
				},
				UpdateIdentifierFunc: func(ctx context.Context, user *system.User) error {
					updated = true
					if user.Email != tt.email {
						t.Errorf("UpdateIdentifier() email = %s, want %s", user.Email, tt.email)
					}
					if user.Phone != tt.phone {
						t.Errorf("UpdateIdentifier() phone = %s, want %s", user.Phone, tt.phone)
					}
					return nil
				},
			}

			svc := &authService{
				userRepo: mockUserRepo,
				totp:     generator,
			}

			errCode, err := svc.UpdateIdentifier(context.Background(), 1, tt.totpCode, tt.password, tt.email, tt.phone)

			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateIdentifier() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if errCode != tt.wantErrCode {
				t.Errorf("UpdateIdentifier() errCode = %v, want %v", errCode, tt.wantErrCode)
			}

			if updated != tt.wantUpdated {
				t.Errorf("UpdateIdentifier() updated = %v, want %v", updated, tt.wantUpdated)
			}
		})
	}
}

func TestAuthService_BindOAuth(t *testing.T) {
	generator := totp.NewGenerator("go-api-admin")
	totpKey := "JBSWY3DPEHPK3PXP"
	validTotpCode := generator.GenerateTOTPCode(totpKey, time.Now())

	tests := []struct {
		name        string
		safeCode    *safeCode
		identifier  string
		password    string
		totpCode    string
		targetUser  *system.User
		bindUser    *system.User
		wantErrCode int
		wantUpdated bool
	}{
		{
			name:        "bind feishu with password",
			safeCode:    &safeCode{Action: "oauth_bind", OauthType: "feishu", OauthID: "ou_xxx"},
			identifier:  "test@example.com",
			password:    "pwd",
			targetUser:  &system.User{Model: gorm.Model{ID: 1}, Status: 1, Salt: "salt", Password: util.MD5("pwd" + "salt")},
			wantErrCode: e.SUCCESS,
			wantUpdated: true,
		},
		{
			name:        "bind wechat with totp",
			safeCode:    &safeCode{Action: "oauth_bind", OauthType: "wechat", OauthID: "wx_xxx"},
			identifier:  "+8613800000000",
			totpCode:    validTotpCode,
			targetUser:  &system.User{Model: gorm.Model{ID: 2}, Status: 1, TotpEnabled: true, TotpKey: totpKey},
			wantErrCode: e.SUCCESS,
			wantUpdated: true,
		},
		{
			name:        "oauth already bound by another user",
			safeCode:    &safeCode{Action: "oauth_bind", OauthType: "feishu", OauthID: "ou_dup"},
			identifier:  "test@example.com",
			password:    "pwd",
			targetUser:  &system.User{Model: gorm.Model{ID: 1}, Status: 1, Salt: "salt", Password: util.MD5("pwd" + "salt")},
			bindUser:    &system.User{Model: gorm.Model{ID: 3}, FeishuId: "ou_dup"},
			wantErrCode: e.IdentifierConflict,
			wantUpdated: false,
		},
		{
			name:        "disabled user can not bind oauth",
			safeCode:    &safeCode{Action: "oauth_bind", OauthType: "feishu", OauthID: "ou_disabled"},
			identifier:  "disabled@example.com",
			password:    "pwd",
			targetUser:  &system.User{Model: gorm.Model{ID: 4}, Status: 2, Salt: "salt", Password: util.MD5("pwd" + "salt")},
			wantErrCode: e.UserNotFound,
			wantUpdated: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updated := false
			mockUserRepo := &mockUserRepo{
				DetailByIdentifierFunc: func(ctx context.Context, identifier string) (*system.User, error) {
					if identifier == tt.identifier {
						return tt.targetUser, nil
					}
					return nil, nil
				},
				DetailFunc: func(ctx context.Context, user *system.User) (*system.User, error) {
					if tt.safeCode != nil && tt.safeCode.OauthType == "feishu" && user.FeishuId == tt.safeCode.OauthID {
						return tt.bindUser, nil
					}
					if tt.safeCode != nil && tt.safeCode.OauthType == "wechat" && user.WechatId == tt.safeCode.OauthID {
						return tt.bindUser, nil
					}
					return nil, nil
				},
				UpdateFunc: func(ctx context.Context, user *system.User) error {
					updated = true
					return nil
				},
			}

			svc := &authService{
				userRepo: mockUserRepo,
				totp:     generator,
			}

			svc.parseSafeCodeFn = func(ctx context.Context, code string) (*safeCode, error) {
				return tt.safeCode, nil
			}

			errCode, _ := svc.BindOAuth(context.Background(), "dummy", tt.identifier, tt.password, tt.totpCode)
			if errCode != tt.wantErrCode {
				t.Errorf("BindOAuth() errCode = %v, want %v", errCode, tt.wantErrCode)
			}
			if updated != tt.wantUpdated {
				t.Errorf("BindOAuth() updated = %v, want %v", updated, tt.wantUpdated)
			}
		})
	}
}
