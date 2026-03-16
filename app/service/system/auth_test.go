package system

import (
	"context"
	"errors"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/go-resty/resty/v2"
	"gorm.io/gorm"

	"github.com/seakee/go-api/app/model/system"
	"github.com/seakee/go-api/app/pkg/e"
	pwd "github.com/seakee/go-api/app/pkg/password"
	"github.com/sk-pkg/util"
)

type roundTripFunc func(req *http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

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

func TestAuthService_getFeishuUserProfile(t *testing.T) {
	tests := []struct {
		name        string
		body        string
		wantUnionID string
		wantProfile *OAuthProfile
		wantErr     bool
	}{
		{
			name:        "parse union_id from user info response",
			body:        `{"code":0,"msg":"success","data":{"union_id":"on-d89jhsdhjsajkda7828enjdj328ydhhw3u43yjhdj","user_id":"5d9bdxxx","name":"zhangsan","avatar_url":"https://example.com/avatar.png"}}`,
			wantUnionID: "on-d89jhsdhjsajkda7828enjdj328ydhhw3u43yjhdj",
			wantProfile: &OAuthProfile{UserName: "zhangsan", Avatar: "https://example.com/avatar.png"},
		},
		{
			name:    "provider error returns error",
			body:    `{"code":999,"msg":"forbidden"}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := resty.New()
			client.SetTransport(roundTripFunc(func(req *http.Request) (*http.Response, error) {
				if req.URL.String() != feishuUserInfoAPI {
					t.Fatalf("request URL = %s, want %s", req.URL.String(), feishuUserInfoAPI)
				}
				if got := req.Header.Get("Authorization"); got != "Bearer test-token" {
					t.Fatalf("Authorization = %s, want %s", got, "Bearer test-token")
				}

				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     make(http.Header),
					Body:       io.NopCloser(strings.NewReader(tt.body)),
				}, nil
			}))

			svc := &authService{request: client}
			unionID, profile, err := svc.getFeishuUserProfile(context.Background(), "test-token")
			if (err != nil) != tt.wantErr {
				t.Fatalf("getFeishuUserProfile() error = %v, wantErr %v", err, tt.wantErr)
			}
			if unionID != tt.wantUnionID {
				t.Fatalf("getFeishuUserProfile() unionID = %s, want %s", unionID, tt.wantUnionID)
			}
			if !reflect.DeepEqual(profile, tt.wantProfile) {
				t.Fatalf("getFeishuUserProfile() profile = %+v, want %+v", profile, tt.wantProfile)
			}
		})
	}
}

func TestAuthService_getWechatUserID(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		wantUserID string
		wantErr    bool
	}{
		{
			name:       "parse userid from oauth user info response",
			body:       `{"errcode":0,"errmsg":"ok","userid":"zhangsan","openid":"openid-ignored"}`,
			wantUserID: "zhangsan",
		},
		{
			name:    "openid only response is rejected",
			body:    `{"errcode":0,"errmsg":"ok","openid":"openid-only"}`,
			wantErr: true,
		},
		{
			name:    "provider error returns error",
			body:    `{"errcode":40029,"errmsg":"invalid code"}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := resty.New()
			client.SetTransport(roundTripFunc(func(req *http.Request) (*http.Response, error) {
				if req.URL.String() != wechatUserInfoAPI+"?access_token=test-access-token&code=test-code" {
					t.Fatalf("request URL = %s", req.URL.String())
				}

				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     make(http.Header),
					Body:       io.NopCloser(strings.NewReader(tt.body)),
				}, nil
			}))

			svc := &authService{request: client}
			userID, err := svc.getWechatUserID(context.Background(), "test-access-token", "test-code")
			if (err != nil) != tt.wantErr {
				t.Fatalf("getWechatUserID() error = %v, wantErr %v", err, tt.wantErr)
			}
			if userID != tt.wantUserID {
				t.Fatalf("getWechatUserID() = %s, want %s", userID, tt.wantUserID)
			}
		})
	}
}

func TestAuthService_getWechatUserProfile(t *testing.T) {
	tests := []struct {
		name        string
		body        string
		wantProfile *OAuthProfile
		wantErr     bool
	}{
		{
			name:        "parse wechat detail profile",
			body:        `{"errcode":0,"errmsg":"ok","name":"lisi","avatar":"https://example.com/wechat-avatar.png"}`,
			wantProfile: &OAuthProfile{UserName: "lisi", Avatar: "https://example.com/wechat-avatar.png"},
		},
		{
			name:    "provider error returns error",
			body:    `{"errcode":60111,"errmsg":"user not found"}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := resty.New()
			client.SetTransport(roundTripFunc(func(req *http.Request) (*http.Response, error) {
				if req.URL.String() != wechatUserDetailAPI+"?access_token=test-access-token&userid=wechat-userid" {
					t.Fatalf("request URL = %s", req.URL.String())
				}

				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     make(http.Header),
					Body:       io.NopCloser(strings.NewReader(tt.body)),
				}, nil
			}))

			svc := &authService{request: client}
			profile, err := svc.getWechatUserProfile(context.Background(), "test-access-token", "wechat-userid")
			if (err != nil) != tt.wantErr {
				t.Fatalf("getWechatUserProfile() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(profile, tt.wantProfile) {
				t.Fatalf("getWechatUserProfile() profile = %+v, want %+v", profile, tt.wantProfile)
			}
		})
	}
}

func TestOAuthLoginStatusErrCode(t *testing.T) {
	tests := []struct {
		name        string
		user        *system.User
		wantErrCode int
	}{
		{
			name:        "nil user means unbound, not error",
			user:        nil,
			wantErrCode: e.SUCCESS,
		},
		{
			name:        "active user can continue oauth login",
			user:        &system.User{Model: gorm.Model{ID: 1}, Status: 1},
			wantErrCode: e.SUCCESS,
		},
		{
			name:        "disabled user is rejected in oauth login",
			user:        &system.User{Model: gorm.Model{ID: 2}, Status: 2},
			wantErrCode: e.UserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := oauthLoginStatusErrCode(tt.user); got != tt.wantErrCode {
				t.Fatalf("oauthLoginStatusErrCode() = %d, want %d", got, tt.wantErrCode)
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

// TestAuthService_UpdatePassword tests update password with the unified reauth ticket.
func TestAuthService_UpdatePassword(t *testing.T) {
	tests := []struct {
		name         string
		reauthTicket string
		password     string
		mockUser     *system.User
		parsedTicket *reauthTicket
		wantErrCode  int
		wantErr      bool
		wantUpdated  bool
		wantConsumed bool
	}{
		{
			name:         "password is required",
			reauthTicket: "reauth-code",
			parsedTicket: &reauthTicket{UserID: 1, Action: reauthActionHighRisk},
			wantErrCode:  e.PasswordCanNotBeNull,
			wantErr:      false,
			wantUpdated:  false,
			wantConsumed: false,
		},
		{
			name:         "invalid reauth ticket is rejected",
			reauthTicket: "reauth-code",
			password:     "new-md5-password",
			wantErrCode:  e.InvalidReauthTicket,
			wantErr:      false,
			wantUpdated:  false,
			wantConsumed: false,
		},
		{
			name:         "user not found after reauth",
			reauthTicket: "reauth-code",
			password:     "new-md5-password",
			parsedTicket: &reauthTicket{UserID: 1, Action: reauthActionHighRisk},
			wantErrCode:  e.UserNotFound,
			wantErr:      false,
			wantUpdated:  false,
			wantConsumed: false,
		},
		{
			name:         "update password succeeds and consumes ticket",
			reauthTicket: "reauth-code",
			password:     "new-md5-password",
			mockUser:     &system.User{Model: gorm.Model{ID: 1}, Status: 1},
			parsedTicket: &reauthTicket{UserID: 1, Action: reauthActionHighRisk},
			wantErrCode:  e.SUCCESS,
			wantErr:      false,
			wantUpdated:  true,
			wantConsumed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updated := false
			consumed := false
			mockUserRepo := &mockUserRepo{
				DetailByIDFunc: func(ctx context.Context, id uint) (*system.User, error) {
					if id != 1 {
						t.Fatalf("DetailByID() id = %d, want 1", id)
					}
					return tt.mockUser, nil
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
				parseReauthTicketFn: func(ctx context.Context, code string) (*reauthTicket, error) {
					if tt.password == "" {
						t.Fatalf("parseReauthTicket() called unexpectedly")
					}
					if code != tt.reauthTicket {
						t.Fatalf("parseReauthTicket() code = %s, want %s", code, tt.reauthTicket)
					}
					return tt.parsedTicket, nil
				},
				consumeReauthTicketFn: func(ctx context.Context, code string) error {
					if code != tt.reauthTicket {
						t.Fatalf("consumeReauthTicket() code = %s, want %s", code, tt.reauthTicket)
					}
					consumed = true
					return nil
				},
			}

			errCode, err := svc.UpdatePassword(context.Background(), 1, tt.reauthTicket, tt.password)

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
			if consumed != tt.wantConsumed {
				t.Errorf("UpdatePassword() consumed = %v, want %v", consumed, tt.wantConsumed)
			}
		})
	}
}

// TestAuthService_UpdateIdentifier tests update identifier with the unified reauth ticket.
func TestAuthService_UpdateIdentifier(t *testing.T) {
	tests := []struct {
		name          string
		reauthTicket  string
		email         string
		phone         string
		mockUser      *system.User
		existingEmail *system.User
		existingPhone *system.User
		parsedTicket  *reauthTicket
		wantErrCode   int
		wantErr       bool
		wantUpdated   bool
		wantConsumed  bool
	}{
		{
			name:         "reauth ticket is required",
			email:        "new@example.com",
			wantErrCode:  e.ReauthTicketCanNotBeNull,
			wantErr:      false,
			wantUpdated:  false,
			wantConsumed: false,
		},
		{
			name:         "invalid reauth ticket is rejected",
			reauthTicket: "reauth-code",
			email:        "new@example.com",
			wantErrCode:  e.InvalidReauthTicket,
			wantErr:      false,
			wantUpdated:  false,
			wantConsumed: false,
		},
		{
			name:         "invalid email format is rejected",
			reauthTicket: "reauth-code",
			email:        "invalid-email",
			wantErrCode:  e.InvalidIdentifier,
			wantErr:      false,
			wantUpdated:  false,
			wantConsumed: false,
		},
		{
			name:         "email exists",
			reauthTicket: "reauth-code",
			email:        "duplicated@example.com",
			mockUser:     &system.User{Model: gorm.Model{ID: 1}, Status: 1},
			existingEmail: &system.User{
				Model: gorm.Model{ID: 2},
				Email: "duplicated@example.com",
			},
			parsedTicket: &reauthTicket{UserID: 1, Action: reauthActionHighRisk},
			wantErrCode:  e.IdentifierExists,
			wantErr:      false,
			wantUpdated:  false,
			wantConsumed: false,
		},
		{
			name:         "update identifier succeeds and consumes ticket",
			reauthTicket: "reauth-code",
			phone:        "+8613800000002",
			mockUser: &system.User{
				Model:  gorm.Model{ID: 1},
				Status: 1,
				Email:  "old@example.com",
			},
			parsedTicket: &reauthTicket{UserID: 1, Action: reauthActionHighRisk},
			wantErrCode:  e.SUCCESS,
			wantErr:      false,
			wantUpdated:  true,
			wantConsumed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updated := false
			consumed := false
			mockUserRepo := &mockUserRepo{
				DetailByIDFunc: func(ctx context.Context, id uint) (*system.User, error) {
					if id != 1 {
						t.Fatalf("DetailByID() id = %d, want 1", id)
					}
					return tt.mockUser, nil
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
				parseReauthTicketFn: func(ctx context.Context, code string) (*reauthTicket, error) {
					if tt.reauthTicket == "" {
						t.Fatalf("parseReauthTicket() called unexpectedly")
					}
					if code != tt.reauthTicket {
						t.Fatalf("parseReauthTicket() code = %s, want %s", code, tt.reauthTicket)
					}
					return tt.parsedTicket, nil
				},
				consumeReauthTicketFn: func(ctx context.Context, code string) error {
					if code != tt.reauthTicket {
						t.Fatalf("consumeReauthTicket() code = %s, want %s", code, tt.reauthTicket)
					}
					consumed = true
					return nil
				},
			}

			errCode, err := svc.UpdateIdentifier(context.Background(), 1, tt.reauthTicket, tt.email, tt.phone)

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
			if consumed != tt.wantConsumed {
				t.Errorf("UpdateIdentifier() consumed = %v, want %v", consumed, tt.wantConsumed)
			}
		})
	}
}

func TestAuthService_verifyByPassword_NoLegacyCompatibility(t *testing.T) {
	defaultCredential := util.MD5(DefaultPassword)
	legacySuffix := "legacy-marker"
	legacyStored := util.MD5(defaultCredential + legacySuffix)

	tests := []struct {
		name        string
		user        *system.User
		password    string
		wantErrCode int
		wantUserNil bool
	}{
		{
			name:        "legacy default password is rejected",
			user:        &system.User{Model: gorm.Model{ID: 1}, Status: 1, Password: legacyStored},
			password:    defaultCredential,
			wantErrCode: e.IdentifierOrPasswordFail,
			wantUserNil: false,
		},
		{
			name:        "legacy non-default password is rejected",
			user:        &system.User{Model: gorm.Model{ID: 2}, Status: 1, Password: util.MD5("other-md5-password" + legacySuffix)},
			password:    "other-md5-password",
			wantErrCode: e.IdentifierOrPasswordFail,
			wantUserNil: false,
		},
		{
			name: "bcrypt default password also requires reset",
			user: func() *system.User {
				hashed, err := pwd.HashCredential(defaultCredential)
				if err != nil {
					t.Fatalf("HashCredential() error = %v", err)
				}
				return &system.User{Model: gorm.Model{ID: 3}, Status: 1, Password: hashed}
			}(),
			password:    defaultCredential,
			wantErrCode: e.NeedResetPWD,
			wantUserNil: false,
		},
		{
			name: "bcrypt non-default password login succeeds",
			user: func() *system.User {
				hashed, err := pwd.HashCredential("bcrypt-md5-password")
				if err != nil {
					t.Fatalf("HashCredential() error = %v", err)
				}
				return &system.User{Model: gorm.Model{ID: 4}, Status: 1, Password: hashed}
			}(),
			password:    "bcrypt-md5-password",
			wantErrCode: e.SUCCESS,
			wantUserNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updated := false
			mockUserRepo := &mockUserRepo{
				DetailByIdentifierFunc: func(ctx context.Context, identifier string) (*system.User, error) {
					return tt.user, nil
				},
				UpdateFunc: func(ctx context.Context, user *system.User) error {
					updated = true
					return nil
				},
			}

			svc := &authService{userRepo: mockUserRepo}
			user, errCode, err := svc.verifyByPassword(context.Background(), "test@example.com", tt.password)
			if err != nil {
				t.Fatalf("verifyByPassword() error = %v", err)
			}
			if errCode != tt.wantErrCode {
				t.Fatalf("verifyByPassword() errCode = %d, want %d", errCode, tt.wantErrCode)
			}
			if (user == nil) != tt.wantUserNil {
				t.Fatalf("verifyByPassword() user nil = %v, want %v", user == nil, tt.wantUserNil)
			}
			if updated {
				t.Fatalf("verifyByPassword() unexpected update call for non-legacy-upgrade flow")
			}
		})
	}
}
