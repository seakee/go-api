package system

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/seakee/go-api/app/config"
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
		name                      string
		identifier                string
		password                  string
		safeCode                  string
		totpCode                  string
		identifierUser            *systemModel.User
		safeCodeUser              *systemModel.User
		parsedSafeCode            *safeCode
		wantErrCode               int
		wantResult                ReauthResult
		wantGeneratedSafeCode     *safeCode
		wantGeneratedReauthTicket *reauthTicket
	}{
		{
			name:                      "password only reauth succeeds",
			identifier:                "test@example.com",
			password:                  "pwd",
			identifierUser:            &systemModel.User{Model: gorm.Model{ID: 1}, Status: 1, Password: passwordHash},
			wantErrCode:               e.SUCCESS,
			wantResult:                ReauthResult{ReauthTicket: "reauth-token"},
			wantGeneratedReauthTicket: &reauthTicket{UserID: 1, Action: "oauth_reauth"},
		},
		{
			name:                  "totp enabled returns safe code challenge from password step",
			identifier:            "test@example.com",
			password:              "pwd",
			totpCode:              validTotpCode,
			identifierUser:        &systemModel.User{Model: gorm.Model{ID: 2}, Status: 1, Password: passwordHash, TotpEnabled: true, TotpKey: totpKey},
			wantErrCode:           e.NeedTfa,
			wantResult:            ReauthResult{SafeCode: "reauth-safe-code"},
			wantGeneratedSafeCode: &safeCode{UserID: 2, Action: "oauth_reauth"},
		},
		{
			name:                      "totp challenge succeeds with safe code",
			safeCode:                  "reauth-safe-code",
			totpCode:                  validTotpCode,
			safeCodeUser:              &systemModel.User{Model: gorm.Model{ID: 3}, Status: 1, Password: passwordHash, TotpEnabled: true, TotpKey: totpKey},
			parsedSafeCode:            &safeCode{UserID: 3, Action: "oauth_reauth"},
			wantErrCode:               e.SUCCESS,
			wantResult:                ReauthResult{ReauthTicket: "reauth-token"},
			wantGeneratedReauthTicket: &reauthTicket{UserID: 3, Action: "oauth_reauth"},
		},
		{
			name:           "totp challenge rejects invalid code",
			safeCode:       "reauth-safe-code",
			totpCode:       "000000",
			safeCodeUser:   &systemModel.User{Model: gorm.Model{ID: 4}, Status: 1, Password: passwordHash, TotpEnabled: true, TotpKey: totpKey},
			parsedSafeCode: &safeCode{UserID: 4, Action: "oauth_reauth"},
			wantErrCode:    e.InvalidTotpCode,
		},
		{
			name:           "totp challenge rejects invalid safe code action",
			safeCode:       "reauth-safe-code",
			totpCode:       validTotpCode,
			parsedSafeCode: &safeCode{UserID: 5, Action: "tfa"},
			wantErrCode:    e.InvalidSafeCode,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			safeCodeCalls := 0
			reauthTicketCalls := 0

			svc := &authService{
				userRepo: &mockUserRepo{
					DetailByIdentifierFunc: func(ctx context.Context, identifier string) (*systemModel.User, error) {
						if tt.identifierUser == nil {
							t.Fatalf("DetailByIdentifier() called unexpectedly")
						}
						if identifier != tt.identifier {
							t.Fatalf("DetailByIdentifier() identifier = %s, want %s", identifier, tt.identifier)
						}
						return tt.identifierUser, nil
					},
					DetailByIDFunc: func(ctx context.Context, id uint) (*systemModel.User, error) {
						if tt.safeCodeUser == nil {
							t.Fatalf("DetailByID() called unexpectedly")
						}
						if id != tt.safeCodeUser.ID {
							t.Fatalf("DetailByID() id = %d, want %d", id, tt.safeCodeUser.ID)
						}
						return tt.safeCodeUser, nil
					},
				},
				totp: generator,
				generateSafeCodeFn: func(ctx context.Context, code safeCode) (string, error) {
					safeCodeCalls++
					if tt.wantGeneratedSafeCode == nil {
						t.Fatalf("generateSafeCode() called unexpectedly")
					}
					if code.UserID != tt.wantGeneratedSafeCode.UserID || code.Action != tt.wantGeneratedSafeCode.Action {
						t.Fatalf("generateSafeCode() payload = %+v, want %+v", code, *tt.wantGeneratedSafeCode)
					}
					return "reauth-safe-code", nil
				},
				parseSafeCodeFn: func(ctx context.Context, code string) (*safeCode, error) {
					if tt.parsedSafeCode == nil {
						t.Fatalf("parseSafeCode() called unexpectedly")
					}
					if code != tt.safeCode {
						t.Fatalf("parseSafeCode() code = %s, want %s", code, tt.safeCode)
					}
					return tt.parsedSafeCode, nil
				},
				generateReauthTicketFn: func(ctx context.Context, ticket reauthTicket) (string, error) {
					reauthTicketCalls++
					if tt.wantGeneratedReauthTicket == nil {
						t.Fatalf("generateReauthTicket() called unexpectedly")
					}
					if ticket.UserID != tt.wantGeneratedReauthTicket.UserID || ticket.Action != tt.wantGeneratedReauthTicket.Action {
						t.Fatalf("generateReauthTicket() payload = %+v, want %+v", ticket, *tt.wantGeneratedReauthTicket)
					}
					return "reauth-token", nil
				},
			}

			result, errCode, err := svc.Reauth(context.Background(), tt.identifier, tt.password, tt.safeCode, tt.totpCode)
			if err != nil {
				t.Fatalf("Reauth() error = %v", err)
			}
			if errCode != tt.wantErrCode {
				t.Fatalf("Reauth() errCode = %d, want %d", errCode, tt.wantErrCode)
			}
			if result != tt.wantResult {
				t.Fatalf("Reauth() result = %+v, want %+v", result, tt.wantResult)
			}
			if safeCodeCalls != btoi(tt.wantGeneratedSafeCode != nil) {
				t.Fatalf("generateSafeCode() call count = %d, want %d", safeCodeCalls, btoi(tt.wantGeneratedSafeCode != nil))
			}
			if reauthTicketCalls != btoi(tt.wantGeneratedReauthTicket != nil) {
				t.Fatalf("generateReauthTicket() call count = %d, want %d", reauthTicketCalls, btoi(tt.wantGeneratedReauthTicket != nil))
			}
		})
	}
}

func btoi(v bool) int {
	if v {
		return 1
	}
	return 0
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
					DetailByIDFunc: func(ctx context.Context, id uint) (*systemModel.User, error) {
						if id != 7 {
							t.Fatalf("DetailByID() id = %d, want 7", id)
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
		name             string
		seedUsers        []systemModel.User
		seedIdentity     []systemModel.UserIdentity
		bindTicket       *bindTicket
		reauthTicket     *reauthTicket
		syncFields       []string
		consumeBindErr   error
		consumeReauthErr error
		wantErr          bool
		wantErrCode      int
		wantTokenUser    string
		wantTokenUserID  uint
		wantConsumeBind  bool
		wantConsumeAuth  bool
		verify           func(t *testing.T, db *gorm.DB)
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
			reauthTicket:    &reauthTicket{UserID: 1, Action: "oauth_reauth"},
			syncFields:      []string{"user_name", "avatar"},
			wantErrCode:     e.SUCCESS,
			wantTokenUser:   "来自飞书",
			wantTokenUserID: 1,
			wantConsumeBind: true,
			wantConsumeAuth: true,
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
			reauthTicket:    &reauthTicket{UserID: 1, Action: "oauth_reauth"},
			wantErrCode:     e.IdentifierConflict,
			wantConsumeBind: false,
			wantConsumeAuth: false,
		},
		{
			name: "bind fails when bind ticket cleanup fails",
			seedUsers: []systemModel.User{
				{Model: gorm.Model{ID: 1}, UserName: "before", Status: 1, Password: "hashed"},
			},
			bindTicket: &bindTicket{
				Provider:        "feishu",
				ProviderTenant:  "tenant-a",
				ProviderSubject: "user-cleanup-bind",
				OAuthProfile:    &OAuthProfile{UserName: "来自飞书"},
			},
			reauthTicket:    &reauthTicket{UserID: 1, Action: "oauth_reauth"},
			consumeBindErr:  errors.New("delete bind ticket failed"),
			wantErr:         true,
			wantErrCode:     e.ERROR,
			wantConsumeBind: true,
			wantConsumeAuth: false,
		},
		{
			name: "bind fails when reauth ticket cleanup fails",
			seedUsers: []systemModel.User{
				{Model: gorm.Model{ID: 1}, UserName: "before", Status: 1, Password: "hashed"},
			},
			bindTicket: &bindTicket{
				Provider:        "feishu",
				ProviderTenant:  "tenant-a",
				ProviderSubject: "user-cleanup-reauth",
				OAuthProfile:    &OAuthProfile{UserName: "来自飞书"},
			},
			reauthTicket:     &reauthTicket{UserID: 1, Action: "oauth_reauth"},
			consumeReauthErr: errors.New("delete reauth ticket failed"),
			wantErr:          true,
			wantErrCode:      e.ERROR,
			wantConsumeBind:  true,
			wantConsumeAuth:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			const tokenExpireIn = int64(7200)
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

			consumedBind := false
			consumedReauth := false
			svc := &authService{
				userRepo: repo.NewUserRepo(db, nil, nil),
				config: config.AdminConfig{
					JwtSecret:     "oauth-bind-secret",
					TokenExpireIn: tokenExpireIn,
				},
				parseBindTicketFn: func(ctx context.Context, code string) (*bindTicket, error) {
					return tt.bindTicket, nil
				},
				parseReauthTicketFn: func(ctx context.Context, code string) (*reauthTicket, error) {
					return tt.reauthTicket, nil
				},
				consumeBindTicketFn: func(ctx context.Context, code string) error {
					if code != "bind-code" {
						t.Fatalf("consumeBindTicket() code = %s, want bind-code", code)
					}
					consumedBind = true
					return tt.consumeBindErr
				},
				consumeReauthTicketFn: func(ctx context.Context, code string) error {
					if code != "reauth-code" {
						t.Fatalf("consumeReauthTicket() code = %s, want reauth-code", code)
					}
					consumedReauth = true
					return tt.consumeReauthErr
				},
			}

			token, errCode, err := svc.ConfirmOAuthBind(context.Background(), "bind-code", "reauth-code", tt.syncFields)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("ConfirmOAuthBind() error = nil, want non-nil")
				}
			} else if err != nil {
				t.Fatalf("ConfirmOAuthBind() error = %v", err)
			}
			if errCode != tt.wantErrCode {
				t.Fatalf("ConfirmOAuthBind() errCode = %d, want %d", errCode, tt.wantErrCode)
			}
			if consumedBind != tt.wantConsumeBind {
				t.Fatalf("consumeBindTicket() called = %v, want %v", consumedBind, tt.wantConsumeBind)
			}
			if consumedReauth != tt.wantConsumeAuth {
				t.Fatalf("consumeReauthTicket() called = %v, want %v", consumedReauth, tt.wantConsumeAuth)
			}
			if tt.wantTokenUser == "" {
				if token.Token != "" || token.ExpireIn != 0 {
					t.Fatalf("ConfirmOAuthBind() token = %+v, want empty", token)
				}
			} else {
				if token.Token == "" {
					t.Fatalf("ConfirmOAuthBind() token is empty")
				}
				if token.ExpireIn != tokenExpireIn {
					t.Fatalf("ConfirmOAuthBind() expires_in = %d, want %d", token.ExpireIn, tokenExpireIn)
				}
				userName, userID, verifyErr := svc.VerifyToken(token.Token)
				if verifyErr != nil {
					t.Fatalf("VerifyToken() error = %v", verifyErr)
				}
				if userName != tt.wantTokenUser {
					t.Fatalf("VerifyToken() userName = %s, want %s", userName, tt.wantTokenUser)
				}
				if userID != tt.wantTokenUserID {
					t.Fatalf("VerifyToken() userID = %d, want %d", userID, tt.wantTokenUserID)
				}
			}

			if tt.verify != nil {
				tt.verify(t, db)
			}
		})
	}
}
