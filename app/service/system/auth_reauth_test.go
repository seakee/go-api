package system

import (
	"context"
	"testing"
	"time"

	systemModel "github.com/seakee/go-api/app/model/system"
	"github.com/seakee/go-api/app/pkg/e"
	pwd "github.com/seakee/go-api/app/pkg/password"
	"github.com/seakee/go-api/app/pkg/totp"
	"gorm.io/gorm"
)

func TestAuthService_ReauthMethods(t *testing.T) {
	svc := &authService{
		userRepo: &mockUserRepo{
			DetailByIDFunc: func(ctx context.Context, id uint) (*systemModel.User, error) {
				if id != 7 {
					t.Fatalf("DetailByID() id = %d, want 7", id)
				}
				return &systemModel.User{
					Model:       gorm.Model{ID: 7},
					Status:      1,
					Password:    "hashed-password",
					TotpEnabled: true,
				}, nil
			},
		},
		passkeyRepo: &mockUserPasskeyRepo{
			CountByUserIDFunc: func(ctx context.Context, userID uint) (int64, error) {
				if userID != 7 {
					t.Fatalf("CountByUserID() userID = %d, want 7", userID)
				}
				return 2, nil
			},
		},
	}

	result, errCode, err := svc.ReauthMethods(context.Background(), 7)
	if err != nil {
		t.Fatalf("ReauthMethods() error = %v", err)
	}
	if errCode != e.SUCCESS {
		t.Fatalf("ReauthMethods() errCode = %d, want %d", errCode, e.SUCCESS)
	}
	if result.DefaultMethod != reauthMethodPasskey {
		t.Fatalf("ReauthMethods() default_method = %s, want %s", result.DefaultMethod, reauthMethodPasskey)
	}
	if len(result.AvailableMethods) != 2 || result.AvailableMethods[0] != reauthMethodPasskey || result.AvailableMethods[1] != reauthMethodPassword {
		t.Fatalf("ReauthMethods() available_methods = %+v, want [passkey password]", result.AvailableMethods)
	}
	if !result.PasswordRequiresTotp || !result.TotpEnabled || result.PasskeyCount != 2 {
		t.Fatalf("ReauthMethods() result = %+v, unexpected flags", result)
	}
}

func TestAuthService_ReauthByPassword(t *testing.T) {
	passwordHash, err := pwd.HashCredential("current-md5-password")
	if err != nil {
		t.Fatalf("HashCredential() error = %v", err)
	}

	tests := []struct {
		name                  string
		user                  *systemModel.User
		wantErrCode           int
		wantResult            ReauthResult
		wantGeneratedSafeCode *safeCode
		wantGeneratedTicket   *reauthTicket
	}{
		{
			name:                "password only verification returns reauth ticket",
			user:                &systemModel.User{Model: gorm.Model{ID: 7}, Status: 1, Password: passwordHash},
			wantErrCode:         e.SUCCESS,
			wantResult:          ReauthResult{ReauthTicket: "reauth-ticket"},
			wantGeneratedTicket: &reauthTicket{UserID: 7, Action: reauthActionHighRisk},
		},
		{
			name:                  "totp enabled user gets safe code",
			user:                  &systemModel.User{Model: gorm.Model{ID: 7}, Status: 1, Password: passwordHash, TotpEnabled: true},
			wantErrCode:           e.NeedTfa,
			wantResult:            ReauthResult{SafeCode: "safe-code"},
			wantGeneratedSafeCode: &safeCode{UserID: 7, Action: reauthActionHighRisk},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			safeCodeCalls := 0
			ticketCalls := 0

			svc := &authService{
				userRepo: &mockUserRepo{
					DetailByIDFunc: func(ctx context.Context, id uint) (*systemModel.User, error) {
						if id != 7 {
							t.Fatalf("DetailByID() id = %d, want 7", id)
						}
						return tt.user, nil
					},
				},
				generateSafeCodeFn: func(ctx context.Context, code safeCode) (string, error) {
					safeCodeCalls++
					if tt.wantGeneratedSafeCode == nil {
						t.Fatalf("generateSafeCode() called unexpectedly")
					}
					if code != *tt.wantGeneratedSafeCode {
						t.Fatalf("generateSafeCode() payload = %+v, want %+v", code, *tt.wantGeneratedSafeCode)
					}
					return "safe-code", nil
				},
				generateReauthTicketFn: func(ctx context.Context, ticket reauthTicket) (string, error) {
					ticketCalls++
					if tt.wantGeneratedTicket == nil {
						t.Fatalf("generateReauthTicket() called unexpectedly")
					}
					if ticket != *tt.wantGeneratedTicket {
						t.Fatalf("generateReauthTicket() payload = %+v, want %+v", ticket, *tt.wantGeneratedTicket)
					}
					return "reauth-ticket", nil
				},
			}

			result, errCode, err := svc.ReauthByPassword(context.Background(), 7, "current-md5-password")
			if err != nil {
				t.Fatalf("ReauthByPassword() error = %v", err)
			}
			if errCode != tt.wantErrCode {
				t.Fatalf("ReauthByPassword() errCode = %d, want %d", errCode, tt.wantErrCode)
			}
			if result != tt.wantResult {
				t.Fatalf("ReauthByPassword() result = %+v, want %+v", result, tt.wantResult)
			}
			if safeCodeCalls != btoi(tt.wantGeneratedSafeCode != nil) {
				t.Fatalf("generateSafeCode() calls = %d, want %d", safeCodeCalls, btoi(tt.wantGeneratedSafeCode != nil))
			}
			if ticketCalls != btoi(tt.wantGeneratedTicket != nil) {
				t.Fatalf("generateReauthTicket() calls = %d, want %d", ticketCalls, btoi(tt.wantGeneratedTicket != nil))
			}
		})
	}
}

func TestAuthService_ReauthByTotp(t *testing.T) {
	generator := totp.NewGenerator("go-api-admin")
	totpKey := "JBSWY3DPEHPK3PXP"
	validTotpCode := generator.GenerateTOTPCode(totpKey, time.Now())

	svc := &authService{
		userRepo: &mockUserRepo{
			DetailByIDFunc: func(ctx context.Context, id uint) (*systemModel.User, error) {
				if id != 7 {
					t.Fatalf("DetailByID() id = %d, want 7", id)
				}
				return &systemModel.User{
					Model:       gorm.Model{ID: 7},
					Status:      1,
					TotpEnabled: true,
					TotpKey:     totpKey,
				}, nil
			},
		},
		totp: generator,
		parseSafeCodeFn: func(ctx context.Context, code string) (*safeCode, error) {
			if code != "safe-code" {
				t.Fatalf("parseSafeCode() code = %s, want safe-code", code)
			}
			return &safeCode{UserID: 7, Action: reauthActionHighRisk}, nil
		},
		generateReauthTicketFn: func(ctx context.Context, ticket reauthTicket) (string, error) {
			if ticket.UserID != 7 || ticket.Action != reauthActionHighRisk {
				t.Fatalf("generateReauthTicket() payload = %+v, want user 7 high risk", ticket)
			}
			return "reauth-ticket", nil
		},
	}

	result, errCode, err := svc.ReauthByTotp(context.Background(), 7, "safe-code", validTotpCode)
	if err != nil {
		t.Fatalf("ReauthByTotp() error = %v", err)
	}
	if errCode != e.SUCCESS {
		t.Fatalf("ReauthByTotp() errCode = %d, want %d", errCode, e.SUCCESS)
	}
	if result.ReauthTicket != "reauth-ticket" {
		t.Fatalf("ReauthByTotp() reauth_ticket = %s, want reauth-ticket", result.ReauthTicket)
	}
}
