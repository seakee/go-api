package system

import (
	"context"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	systemModel "github.com/seakee/go-api/app/model/system"
	"github.com/seakee/go-api/app/pkg/e"
	pwd "github.com/seakee/go-api/app/pkg/password"
	"github.com/seakee/go-api/app/pkg/totp"
	repo "github.com/seakee/go-api/app/repository/system"
	"gorm.io/gorm"
)

func TestAuthService_Reauth(t *testing.T) {
	generator := totp.NewGenerator("go-api-admin")
	totpKey := "JBSWY3DPEHPK3PXP"
	validTotpCode := generator.GenerateTOTPCode(totpKey, time.Now())
	passwordHash, err := pwd.HashCredential("pwd")
	if err != nil {
		t.Fatalf("HashCredential() error = %v", err)
	}

	tests := []struct {
		name        string
		user        *systemModel.User
		totpCode    string
		wantErrCode int
		wantTicket  string
	}{
		{
			name:        "password only reauth succeeds",
			user:        &systemModel.User{Model: gorm.Model{ID: 1}, Status: 1, Password: passwordHash},
			wantErrCode: e.SUCCESS,
			wantTicket:  "reauth-token",
		},
		{
			name:        "totp enabled requires valid code",
			user:        &systemModel.User{Model: gorm.Model{ID: 2}, Status: 1, Password: passwordHash, TotpEnabled: true, TotpKey: totpKey},
			totpCode:    validTotpCode,
			wantErrCode: e.SUCCESS,
			wantTicket:  "reauth-token",
		},
		{
			name:        "totp enabled rejects empty code",
			user:        &systemModel.User{Model: gorm.Model{ID: 3}, Status: 1, Password: passwordHash, TotpEnabled: true, TotpKey: totpKey},
			wantErrCode: e.TotpCodeCanNotBeNull,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &authService{
				userRepo: &mockUserRepo{
					DetailByIdentifierFunc: func(ctx context.Context, identifier string) (*systemModel.User, error) {
						return tt.user, nil
					},
				},
				totp: generator,
				generateReauthTicketFn: func(ctx context.Context, ticket reauthTicket) (string, error) {
					if ticket.UserID != tt.user.ID {
						t.Fatalf("generateReauthTicket() userID = %d, want %d", ticket.UserID, tt.user.ID)
					}
					if ticket.Action != "oauth_reauth" {
						t.Fatalf("generateReauthTicket() action = %s, want oauth_reauth", ticket.Action)
					}
					return "reauth-token", nil
				},
			}

			ticket, errCode, err := svc.Reauth(context.Background(), "test@example.com", "pwd", tt.totpCode)
			if err != nil {
				t.Fatalf("Reauth() error = %v", err)
			}
			if errCode != tt.wantErrCode {
				t.Fatalf("Reauth() errCode = %d, want %d", errCode, tt.wantErrCode)
			}
			if ticket != tt.wantTicket {
				t.Fatalf("Reauth() ticket = %s, want %s", ticket, tt.wantTicket)
			}
		})
	}
}

func TestAuthService_UnbindOAuth(t *testing.T) {
	tests := []struct {
		name            string
		user            *systemModel.User
		targetIdentity  *systemModel.UserIdentity
		identityList    []systemModel.UserIdentity
		wantErrCode     int
		wantDeleteCalls int
		wantConsume     bool
	}{
		{
			name:           "unbind removes only specified identity",
			user:           &systemModel.User{Model: gorm.Model{ID: 7}, Status: 1, Password: "hashed"},
			targetIdentity: &systemModel.UserIdentity{Model: gorm.Model{ID: 12}, UserID: 7, Provider: "feishu", ProviderTenant: "tenant-a"},
			identityList: []systemModel.UserIdentity{
				{Model: gorm.Model{ID: 12}, UserID: 7, Provider: "feishu", ProviderTenant: "tenant-a"},
				{Model: gorm.Model{ID: 13}, UserID: 7, Provider: "feishu", ProviderTenant: "tenant-b"},
			},
			wantErrCode:     e.SUCCESS,
			wantDeleteCalls: 1,
			wantConsume:     true,
		},
		{
			name:           "identity not bound returns explicit error",
			user:           &systemModel.User{Model: gorm.Model{ID: 7}, Status: 1, Password: "hashed"},
			targetIdentity: nil,
			wantErrCode:    e.OauthAccountNotBound,
		},
		{
			name:           "last login method cannot be removed",
			user:           &systemModel.User{Model: gorm.Model{ID: 7}, Status: 1},
			targetIdentity: &systemModel.UserIdentity{Model: gorm.Model{ID: 12}, UserID: 7, Provider: "feishu", ProviderTenant: "tenant-a"},
			identityList: []systemModel.UserIdentity{
				{Model: gorm.Model{ID: 12}, UserID: 7, Provider: "feishu", ProviderTenant: "tenant-a"},
			},
			wantErrCode: e.LastLoginMethodCannotBeRemoved,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deleteCalls := 0
			consumed := false

			svc := &authService{
				userRepo: &mockUserRepo{
					DetailFunc: func(ctx context.Context, user *systemModel.User) (*systemModel.User, error) {
						if user.ID != 7 {
							t.Fatalf("Detail() user id = %d, want 7", user.ID)
						}
						return tt.user, nil
					},
				},
				identityRepo: &mockUserIdentityRepo{
					DetailByIDAndUserIDFunc: func(ctx context.Context, id, userID uint) (*systemModel.UserIdentity, error) {
						if id != 12 {
							t.Fatalf("DetailByIDAndUserID() id = %d, want 12", id)
						}
						if userID != 7 {
							t.Fatalf("DetailByIDAndUserID() userID = %d, want 7", userID)
						}
						return tt.targetIdentity, nil
					},
					ListByUserIDFunc: func(ctx context.Context, userID uint) ([]systemModel.UserIdentity, error) {
						if userID != 7 {
							t.Fatalf("ListByUserID() userID = %d, want 7", userID)
						}
						return tt.identityList, nil
					},
					DeleteByIDAndUserIDFunc: func(ctx context.Context, id, userID uint) error {
						deleteCalls++
						if id != 12 {
							t.Fatalf("DeleteByIDAndUserID() id = %d, want 12", id)
						}
						if userID != 7 {
							t.Fatalf("DeleteByIDAndUserID() userID = %d, want 7", userID)
						}
						return nil
					},
				},
				parseReauthTicketFn: func(ctx context.Context, code string) (*reauthTicket, error) {
					if code != "reauth-code" {
						t.Fatalf("parseReauthTicket() code = %s, want reauth-code", code)
					}
					return &reauthTicket{UserID: 7, Action: "oauth_reauth"}, nil
				},
				consumeReauthTicketFn: func(ctx context.Context, code string) error {
					if code != "reauth-code" {
						t.Fatalf("consumeReauthTicket() code = %s, want reauth-code", code)
					}
					consumed = true
					return nil
				},
			}

			errCode, err := svc.UnbindOAuth(context.Background(), 7, 12, "reauth-code")
			if err != nil {
				t.Fatalf("UnbindOAuth() error = %v", err)
			}
			if errCode != tt.wantErrCode {
				t.Fatalf("UnbindOAuth() errCode = %d, want %d", errCode, tt.wantErrCode)
			}
			if deleteCalls != tt.wantDeleteCalls {
				t.Fatalf("DeleteByIDAndUserID() call count = %d, want %d", deleteCalls, tt.wantDeleteCalls)
			}
			if consumed != tt.wantConsume {
				t.Fatalf("consumeReauthTicket() called = %v, want %v", consumed, tt.wantConsume)
			}
		})
	}
}

func TestAuthService_ConfirmOAuthBind(t *testing.T) {
	tests := []struct {
		name         string
		seedUsers    []systemModel.User
		seedIdentity []systemModel.UserIdentity
		bindTicket   *bindTicket
		reauthTicket *reauthTicket
		syncFields   []string
		wantErrCode  int
		verify       func(t *testing.T, db *gorm.DB)
	}{
		{
			name: "bind creates identity and syncs profile",
			seedUsers: []systemModel.User{
				{Model: gorm.Model{ID: 1}, UserName: "before", Status: 1, Password: "hashed"},
			},
			bindTicket: &bindTicket{
				Provider:        "feishu",
				ProviderTenant:  "tenant-a",
				ProviderSubject: "user-001",
				OAuthProfile:    &OAuthProfile{UserName: "来自飞书", Avatar: "https://example.com/avatar.png"},
			},
			reauthTicket: &reauthTicket{UserID: 1, Action: "oauth_reauth"},
			syncFields:   []string{"user_name", "avatar"},
			wantErrCode:  e.SUCCESS,
			verify: func(t *testing.T, db *gorm.DB) {
				t.Helper()

				var user systemModel.User
				if err := db.First(&user, 1).Error; err != nil {
					t.Fatalf("query user error = %v", err)
				}
				if user.UserName != "来自飞书" {
					t.Fatalf("user.UserName = %s, want 来自飞书", user.UserName)
				}
				if user.Avatar != "https://example.com/avatar.png" {
					t.Fatalf("user.Avatar = %s, want synced avatar", user.Avatar)
				}

				identityRepo := repo.NewUserIdentityRepo(db, nil, nil)
				identity, err := identityRepo.DetailByProvider(context.Background(), "feishu", "tenant-a", "user-001")
				if err != nil {
					t.Fatalf("DetailByProvider() error = %v", err)
				}
				if identity == nil {
					t.Fatalf("DetailByProvider() identity = nil, want record")
				}
				if identity.UserID != 1 {
					t.Fatalf("identity.UserID = %d, want 1", identity.UserID)
				}
			},
		},
		{
			name: "bind rejects identity already bound by another user",
			seedUsers: []systemModel.User{
				{Model: gorm.Model{ID: 1}, UserName: "target", Status: 1, Password: "hashed"},
				{Model: gorm.Model{ID: 2}, UserName: "other", Status: 1, Password: "hashed"},
			},
			seedIdentity: []systemModel.UserIdentity{
				{UserID: 2, Provider: "feishu", ProviderTenant: "tenant-a", ProviderSubject: "user-dup"},
			},
			bindTicket: &bindTicket{
				Provider:        "feishu",
				ProviderTenant:  "tenant-a",
				ProviderSubject: "user-dup",
			},
			reauthTicket: &reauthTicket{UserID: 1, Action: "oauth_reauth"},
			wantErrCode:  e.IdentifierConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
			if err != nil {
				t.Fatalf("gorm.Open() error = %v", err)
			}
			if err := db.AutoMigrate(&systemModel.User{}, &systemModel.UserIdentity{}); err != nil {
				t.Fatalf("AutoMigrate() error = %v", err)
			}
			if len(tt.seedUsers) > 0 {
				if err := db.Create(&tt.seedUsers).Error; err != nil {
					t.Fatalf("seed users error = %v", err)
				}
			}
			if len(tt.seedIdentity) > 0 {
				if err := db.Create(&tt.seedIdentity).Error; err != nil {
					t.Fatalf("seed identities error = %v", err)
				}
			}

			svc := &authService{
				db: db,
				parseBindTicketFn: func(ctx context.Context, code string) (*bindTicket, error) {
					return tt.bindTicket, nil
				},
				parseReauthTicketFn: func(ctx context.Context, code string) (*reauthTicket, error) {
					return tt.reauthTicket, nil
				},
				consumeBindTicketFn:   func(ctx context.Context, code string) error { return nil },
				consumeReauthTicketFn: func(ctx context.Context, code string) error { return nil },
			}

			errCode, err := svc.ConfirmOAuthBind(context.Background(), "bind-code", "reauth-code", tt.syncFields)
			if err != nil {
				t.Fatalf("ConfirmOAuthBind() error = %v", err)
			}
			if errCode != tt.wantErrCode {
				t.Fatalf("ConfirmOAuthBind() errCode = %d, want %d", errCode, tt.wantErrCode)
			}

			if tt.verify != nil {
				tt.verify(t, db)
			}
		})
	}
}
