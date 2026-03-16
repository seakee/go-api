package system

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"github.com/seakee/go-api/app/model/system"
	roleModel "github.com/seakee/go-api/app/model/system/role"
	"github.com/seakee/go-api/app/pkg/e"
	repo "github.com/seakee/go-api/app/repository/system"
)

// TestUserService_Detail tests fetching user details.
func TestUserService_Detail(t *testing.T) {
	tests := []struct {
		name        string
		userID      uint
		mockUser    *system.User
		mockErr     error
		wantErrCode int
		wantErr     bool
	}{
		{
			name:   "get user detail successfully",
			userID: 1,
			mockUser: &system.User{
				Model:    gorm.Model{ID: 1},
				Email:    "test@example.com",
				Phone:    "+8613800000000",
				UserName: "Test User",
			},
			mockErr:     nil,
			wantErrCode: e.SUCCESS,
			wantErr:     false,
		},
		{
			name:        "user not found",
			userID:      999,
			mockUser:    nil,
			mockErr:     errors.New("record not found"),
			wantErrCode: e.UserNotFound,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserRepo := &mockUserRepo{
				DetailFunc: func(ctx context.Context, user *system.User) (*system.User, error) {
					if user.ID == tt.userID {
						return tt.mockUser, tt.mockErr
					}
					return nil, errors.New("record not found")
				},
			}

			svc := &userService{
				userRepo: mockUserRepo,
			}

			user, errCode, err := svc.Detail(context.Background(), tt.userID)

			if (err != nil) != tt.wantErr {
				t.Errorf("Detail() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if errCode != tt.wantErrCode {
				t.Errorf("Detail() errCode = %v, want %v", errCode, tt.wantErrCode)
				return
			}
			if !tt.wantErr && user == nil {
				t.Errorf("Detail() user should not be nil")
			}
		})
	}
}

// TestUserService_Paginate tests user pagination.
func TestUserService_Paginate(t *testing.T) {
	tests := []struct {
		name      string
		page      int
		size      int
		mockUsers []system.User
		mockCount int64
		mockErr   error
		wantTotal int64
		wantLen   int
		wantErr   bool
	}{
		{
			name: "get user list successfully",
			page: 1,
			size: 10,
			mockUsers: []system.User{
				{Model: gorm.Model{ID: 1}, Email: "user1@example.com", Phone: "+8613800000001"},
				{Model: gorm.Model{ID: 2}, Email: "user2@example.com", Phone: "+8613800000002"},
			},
			mockCount: 2,
			mockErr:   nil,
			wantTotal: 2,
			wantLen:   2,
			wantErr:   false,
		},
		{
			name:      "empty list",
			page:      1,
			size:      10,
			mockUsers: []system.User{},
			mockCount: 0,
			mockErr:   nil,
			wantTotal: 0,
			wantLen:   0,
			wantErr:   false,
		},
		{
			name:      "query error",
			page:      1,
			size:      10,
			mockUsers: nil,
			mockCount: 0,
			mockErr:   errors.New("database error"),
			wantTotal: 0,
			wantLen:   0,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserRepo := &mockUserRepo{
				CountFunc: func(ctx context.Context, user *system.User) (int64, error) {
					if tt.mockErr != nil {
						return 0, tt.mockErr
					}
					return tt.mockCount, nil
				},
				PaginateFunc: func(ctx context.Context, user *system.User, page, pageSize int) ([]system.User, error) {
					return tt.mockUsers, nil
				},
			}

			svc := &userService{
				userRepo: mockUserRepo,
			}

			list, total, err := svc.Paginate(context.Background(), &system.User{}, tt.page, tt.size)

			if (err != nil) != tt.wantErr {
				t.Errorf("Paginate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if total != tt.wantTotal {
				t.Errorf("Paginate() total = %v, want %v", total, tt.wantTotal)
			}
			if len(list) != tt.wantLen {
				t.Errorf("Paginate() len = %v, want %v", len(list), tt.wantLen)
			}
		})
	}
}

// Note: Roles, UpdateRole, and Delete depend on the RoleUserRepo interface,
// which includes an unexported i() method to prevent external implementations.
// Tests for these methods require integration tests or repository-level test support.

func TestUserService_ListPasskeys(t *testing.T) {
	t.Run("rejects empty user id", func(t *testing.T) {
		svc := &userService{}

		list, errCode, err := svc.ListPasskeys(context.Background(), 0)
		if err != nil {
			t.Fatalf("ListPasskeys() error = %v, want nil", err)
		}
		if errCode != e.InvalidParams {
			t.Fatalf("ListPasskeys() errCode = %d, want %d", errCode, e.InvalidParams)
		}
		if list != nil {
			t.Fatalf("ListPasskeys() list = %+v, want nil", list)
		}
	})

	t.Run("returns user not found", func(t *testing.T) {
		svc := &userService{
			userRepo: &mockUserRepo{
				DetailByIDFunc: func(ctx context.Context, id uint) (*system.User, error) {
					return nil, nil
				},
			},
			passkeyRepo: &mockUserPasskeyRepo{},
		}

		list, errCode, err := svc.ListPasskeys(context.Background(), 1)
		if err != nil {
			t.Fatalf("ListPasskeys() error = %v, want nil", err)
		}
		if errCode != e.UserNotFound {
			t.Fatalf("ListPasskeys() errCode = %d, want %d", errCode, e.UserNotFound)
		}
		if list != nil {
			t.Fatalf("ListPasskeys() list = %+v, want nil", list)
		}
	})

	t.Run("returns mapped passkey list", func(t *testing.T) {
		usedAt := time.Unix(1700000000, 0).UTC()
		svc := &userService{
			userRepo: &mockUserRepo{
				DetailByIDFunc: func(ctx context.Context, id uint) (*system.User, error) {
					return &system.User{Model: gorm.Model{ID: id}, UserName: "u1"}, nil
				},
			},
			passkeyRepo: &mockUserPasskeyRepo{
				ListByUserIDFunc: func(ctx context.Context, userID uint) ([]system.UserPasskey, error) {
					return []system.UserPasskey{
						{
							Model:          gorm.Model{ID: 11, CreatedAt: time.Unix(1700000100, 0).UTC()},
							UserID:         userID,
							DisplayName:    "MacBook",
							AAGUID:         "aaguid-1",
							TransportsJSON: `["internal","hybrid"]`,
							LastUsedAt:     &usedAt,
						},
					}, nil
				},
			},
		}

		list, errCode, err := svc.ListPasskeys(context.Background(), 1)
		if err != nil {
			t.Fatalf("ListPasskeys() error = %v", err)
		}
		if errCode != e.SUCCESS {
			t.Fatalf("ListPasskeys() errCode = %d, want %d", errCode, e.SUCCESS)
		}
		if len(list) != 1 {
			t.Fatalf("ListPasskeys() len = %d, want 1", len(list))
		}
		if list[0].ID != 11 || list[0].DisplayName != "MacBook" || list[0].AAGUID != "aaguid-1" {
			t.Fatalf("ListPasskeys() item = %+v, unexpected fields", list[0])
		}
		if len(list[0].Transports) != 2 || list[0].Transports[0] != "internal" || list[0].Transports[1] != "hybrid" {
			t.Fatalf("ListPasskeys() transports = %+v, want [internal hybrid]", list[0].Transports)
		}
		if list[0].LastUsedAt == nil || !list[0].LastUsedAt.UTC().Equal(usedAt) {
			t.Fatalf("ListPasskeys() LastUsedAt = %v, want %v", list[0].LastUsedAt, usedAt)
		}
	})
}

func TestUserService_DeletePasskey(t *testing.T) {
	t.Run("rejects invalid params", func(t *testing.T) {
		svc := &userService{}

		errCode, err := svc.DeletePasskey(context.Background(), 7, 0, 1, "reauth-code")
		if err != nil {
			t.Fatalf("DeletePasskey() error = %v, want nil", err)
		}
		if errCode != e.InvalidParams {
			t.Fatalf("DeletePasskey() errCode = %d, want %d", errCode, e.InvalidParams)
		}
	})

	t.Run("missing reauth ticket is rejected", func(t *testing.T) {
		svc := &userService{}

		errCode, err := svc.DeletePasskey(context.Background(), 7, 1, 2, "")
		if err != nil {
			t.Fatalf("DeletePasskey() error = %v, want nil", err)
		}
		if errCode != e.ReauthTicketCanNotBeNull {
			t.Fatalf("DeletePasskey() errCode = %d, want %d", errCode, e.ReauthTicketCanNotBeNull)
		}
	})

	t.Run("rejects super admin operation", func(t *testing.T) {
		svc := &userService{
			parseReauthTicketFn: func(ctx context.Context, code string) (*reauthTicket, error) {
				return &reauthTicket{UserID: 7, Action: reauthActionHighRisk}, nil
			},
			authRepo: &mockAuthRepo{
				HasRoleFunc: func(ctx context.Context, userID uint, role string) (bool, error) {
					return true, nil
				},
			},
		}

		errCode, err := svc.DeletePasskey(context.Background(), 7, 1, 2, "reauth-code")
		if err == nil {
			t.Fatalf("DeletePasskey() error = nil, want non-nil")
		}
		if errCode != e.UserCanNotBeOperated {
			t.Fatalf("DeletePasskey() errCode = %d, want %d", errCode, e.UserCanNotBeOperated)
		}
	})

	t.Run("returns not found when passkey missing", func(t *testing.T) {
		svc := &userService{
			parseReauthTicketFn: func(ctx context.Context, code string) (*reauthTicket, error) {
				return &reauthTicket{UserID: 7, Action: reauthActionHighRisk}, nil
			},
			authRepo: &mockAuthRepo{
				HasRoleFunc: func(ctx context.Context, userID uint, role string) (bool, error) {
					return false, nil
				},
			},
			passkeyRepo: &mockUserPasskeyRepo{
				DetailByIDAndUserIDFunc: func(ctx context.Context, id, userID uint) (*system.UserPasskey, error) {
					return nil, nil
				},
			},
		}

		errCode, err := svc.DeletePasskey(context.Background(), 7, 1, 2, "reauth-code")
		if err != nil {
			t.Fatalf("DeletePasskey() error = %v, want nil", err)
		}
		if errCode != e.PasskeyCredentialNotFound {
			t.Fatalf("DeletePasskey() errCode = %d, want %d", errCode, e.PasskeyCredentialNotFound)
		}
	})

	t.Run("rejects removing last login method", func(t *testing.T) {
		deleteCalled := false
		consumed := false
		svc := &userService{
			parseReauthTicketFn: func(ctx context.Context, code string) (*reauthTicket, error) {
				return &reauthTicket{UserID: 7, Action: reauthActionHighRisk}, nil
			},
			consumeReauthTicketFn: func(ctx context.Context, code string) error {
				consumed = true
				return nil
			},
			authRepo: &mockAuthRepo{
				HasRoleFunc: func(ctx context.Context, userID uint, role string) (bool, error) {
					return false, nil
				},
			},
			userRepo: &mockUserRepo{
				DetailByIDFunc: func(ctx context.Context, id uint) (*system.User, error) {
					return &system.User{Model: gorm.Model{ID: id}}, nil
				},
			},
			passkeyRepo: &mockUserPasskeyRepo{
				DetailByIDAndUserIDFunc: func(ctx context.Context, id, userID uint) (*system.UserPasskey, error) {
					return &system.UserPasskey{Model: gorm.Model{ID: id}, UserID: userID}, nil
				},
				CountByUserIDFunc: func(ctx context.Context, userID uint) (int64, error) {
					return 1, nil
				},
				DeleteByIDAndUserIDFunc: func(ctx context.Context, id, userID uint) error {
					deleteCalled = true
					return nil
				},
			},
		}

		errCode, err := svc.DeletePasskey(context.Background(), 7, 1, 2, "reauth-code")
		if err != nil {
			t.Fatalf("DeletePasskey() error = %v, want nil", err)
		}
		if errCode != e.LastLoginMethodCannotBeRemoved {
			t.Fatalf("DeletePasskey() errCode = %d, want %d", errCode, e.LastLoginMethodCannotBeRemoved)
		}
		if deleteCalled {
			t.Fatalf("DeletePasskey() should not call delete")
		}
		if consumed {
			t.Fatalf("DeletePasskey() should not consume ticket on failure")
		}
	})

	t.Run("deletes passkey when other method exists", func(t *testing.T) {
		deleteCalled := false
		consumed := false
		svc := &userService{
			parseReauthTicketFn: func(ctx context.Context, code string) (*reauthTicket, error) {
				return &reauthTicket{UserID: 7, Action: reauthActionHighRisk}, nil
			},
			consumeReauthTicketFn: func(ctx context.Context, code string) error {
				consumed = true
				return nil
			},
			authRepo: &mockAuthRepo{
				HasRoleFunc: func(ctx context.Context, userID uint, role string) (bool, error) {
					return false, nil
				},
			},
			userRepo: &mockUserRepo{
				DetailByIDFunc: func(ctx context.Context, id uint) (*system.User, error) {
					return &system.User{Model: gorm.Model{ID: id}, Password: "hashed"}, nil
				},
			},
			passkeyRepo: &mockUserPasskeyRepo{
				DetailByIDAndUserIDFunc: func(ctx context.Context, id, userID uint) (*system.UserPasskey, error) {
					return &system.UserPasskey{Model: gorm.Model{ID: id}, UserID: userID}, nil
				},
				CountByUserIDFunc: func(ctx context.Context, userID uint) (int64, error) {
					return 1, nil
				},
				DeleteByIDAndUserIDFunc: func(ctx context.Context, id, userID uint) error {
					deleteCalled = true
					return nil
				},
			},
		}

		errCode, err := svc.DeletePasskey(context.Background(), 7, 1, 2, "reauth-code")
		if err != nil {
			t.Fatalf("DeletePasskey() error = %v", err)
		}
		if errCode != e.SUCCESS {
			t.Fatalf("DeletePasskey() errCode = %d, want %d", errCode, e.SUCCESS)
		}
		if !deleteCalled {
			t.Fatalf("DeletePasskey() delete was not called")
		}
		if !consumed {
			t.Fatalf("DeletePasskey() should consume ticket on success")
		}
	})

	t.Run("deletes passkey when another passkey remains", func(t *testing.T) {
		deleteCalled := false
		consumed := false
		svc := &userService{
			parseReauthTicketFn: func(ctx context.Context, code string) (*reauthTicket, error) {
				return &reauthTicket{UserID: 7, Action: reauthActionHighRisk}, nil
			},
			consumeReauthTicketFn: func(ctx context.Context, code string) error {
				consumed = true
				return nil
			},
			authRepo: &mockAuthRepo{
				HasRoleFunc: func(ctx context.Context, userID uint, role string) (bool, error) {
					return false, nil
				},
			},
			userRepo: &mockUserRepo{
				DetailByIDFunc: func(ctx context.Context, id uint) (*system.User, error) {
					return &system.User{Model: gorm.Model{ID: id}}, nil
				},
			},
			passkeyRepo: &mockUserPasskeyRepo{
				DetailByIDAndUserIDFunc: func(ctx context.Context, id, userID uint) (*system.UserPasskey, error) {
					return &system.UserPasskey{Model: gorm.Model{ID: id}, UserID: userID}, nil
				},
				CountByUserIDFunc: func(ctx context.Context, userID uint) (int64, error) {
					return 2, nil
				},
				DeleteByIDAndUserIDFunc: func(ctx context.Context, id, userID uint) error {
					deleteCalled = true
					return nil
				},
			},
		}

		errCode, err := svc.DeletePasskey(context.Background(), 7, 1, 2, "reauth-code")
		if err != nil {
			t.Fatalf("DeletePasskey() error = %v", err)
		}
		if errCode != e.SUCCESS {
			t.Fatalf("DeletePasskey() errCode = %d, want %d", errCode, e.SUCCESS)
		}
		if !deleteCalled {
			t.Fatalf("DeletePasskey() delete was not called")
		}
		if !consumed {
			t.Fatalf("DeletePasskey() should consume ticket on success")
		}
	})
}

func TestUserService_DeleteAllPasskeys(t *testing.T) {
	t.Run("rejects invalid user id", func(t *testing.T) {
		svc := &userService{}

		errCode, err := svc.DeleteAllPasskeys(context.Background(), 7, 0, "reauth-code")
		if err != nil {
			t.Fatalf("DeleteAllPasskeys() error = %v, want nil", err)
		}
		if errCode != e.InvalidParams {
			t.Fatalf("DeleteAllPasskeys() errCode = %d, want %d", errCode, e.InvalidParams)
		}
	})

	t.Run("invalid reauth ticket is rejected", func(t *testing.T) {
		svc := &userService{}

		errCode, err := svc.DeleteAllPasskeys(context.Background(), 7, 1, "reauth-code")
		if err != nil {
			t.Fatalf("DeleteAllPasskeys() error = %v, want nil", err)
		}
		if errCode != e.InvalidReauthTicket {
			t.Fatalf("DeleteAllPasskeys() errCode = %d, want %d", errCode, e.InvalidReauthTicket)
		}
	})

	t.Run("returns success when user has no passkeys", func(t *testing.T) {
		deleteCalled := false
		svc := &userService{
			parseReauthTicketFn: func(ctx context.Context, code string) (*reauthTicket, error) {
				return &reauthTicket{UserID: 7, Action: reauthActionHighRisk}, nil
			},
			authRepo: &mockAuthRepo{
				HasRoleFunc: func(ctx context.Context, userID uint, role string) (bool, error) {
					return false, nil
				},
			},
			passkeyRepo: &mockUserPasskeyRepo{
				CountByUserIDFunc: func(ctx context.Context, userID uint) (int64, error) {
					return 0, nil
				},
				DeleteByUserIDFunc: func(ctx context.Context, userID uint) error {
					deleteCalled = true
					return nil
				},
			},
		}

		errCode, err := svc.DeleteAllPasskeys(context.Background(), 7, 1, "reauth-code")
		if err != nil {
			t.Fatalf("DeleteAllPasskeys() error = %v", err)
		}
		if errCode != e.SUCCESS {
			t.Fatalf("DeleteAllPasskeys() errCode = %d, want %d", errCode, e.SUCCESS)
		}
		if deleteCalled {
			t.Fatalf("DeleteAllPasskeys() should not call delete")
		}
	})

	t.Run("rejects removing last login method", func(t *testing.T) {
		svc := &userService{
			parseReauthTicketFn: func(ctx context.Context, code string) (*reauthTicket, error) {
				return &reauthTicket{UserID: 7, Action: reauthActionHighRisk}, nil
			},
			authRepo: &mockAuthRepo{
				HasRoleFunc: func(ctx context.Context, userID uint, role string) (bool, error) {
					return false, nil
				},
			},
			userRepo: &mockUserRepo{
				DetailByIDFunc: func(ctx context.Context, id uint) (*system.User, error) {
					return &system.User{Model: gorm.Model{ID: id}}, nil
				},
			},
			passkeyRepo: &mockUserPasskeyRepo{
				CountByUserIDFunc: func(ctx context.Context, userID uint) (int64, error) {
					return 2, nil
				},
			},
		}

		errCode, err := svc.DeleteAllPasskeys(context.Background(), 7, 1, "reauth-code")
		if err != nil {
			t.Fatalf("DeleteAllPasskeys() error = %v", err)
		}
		if errCode != e.LastLoginMethodCannotBeRemoved {
			t.Fatalf("DeleteAllPasskeys() errCode = %d, want %d", errCode, e.LastLoginMethodCannotBeRemoved)
		}
	})

	t.Run("deletes all passkeys when other method exists", func(t *testing.T) {
		deleteCalled := false
		consumed := false
		svc := &userService{
			parseReauthTicketFn: func(ctx context.Context, code string) (*reauthTicket, error) {
				return &reauthTicket{UserID: 7, Action: reauthActionHighRisk}, nil
			},
			consumeReauthTicketFn: func(ctx context.Context, code string) error {
				consumed = true
				return nil
			},
			authRepo: &mockAuthRepo{
				HasRoleFunc: func(ctx context.Context, userID uint, role string) (bool, error) {
					return false, nil
				},
			},
			userRepo: &mockUserRepo{
				DetailByIDFunc: func(ctx context.Context, id uint) (*system.User, error) {
					return &system.User{Model: gorm.Model{ID: id}, Password: "hashed"}, nil
				},
			},
			passkeyRepo: &mockUserPasskeyRepo{
				CountByUserIDFunc: func(ctx context.Context, userID uint) (int64, error) {
					return 2, nil
				},
				DeleteByUserIDFunc: func(ctx context.Context, userID uint) error {
					deleteCalled = true
					return nil
				},
			},
		}

		errCode, err := svc.DeleteAllPasskeys(context.Background(), 7, 1, "reauth-code")
		if err != nil {
			t.Fatalf("DeleteAllPasskeys() error = %v", err)
		}
		if errCode != e.SUCCESS {
			t.Fatalf("DeleteAllPasskeys() errCode = %d, want %d", errCode, e.SUCCESS)
		}
		if !deleteCalled {
			t.Fatalf("DeleteAllPasskeys() delete was not called")
		}
		if !consumed {
			t.Fatalf("DeleteAllPasskeys() should consume ticket on success")
		}
	})
}

func TestUserService_Delete(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open() error = %v", err)
	}

	if err := db.AutoMigrate(&system.User{}, &system.UserIdentity{}, &system.UserPasskey{}, &roleModel.User{}); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}

	users := []system.User{
		{Model: gorm.Model{ID: 1}, UserName: "delete-me", Status: 1},
		{Model: gorm.Model{ID: 2}, UserName: "keep-me", Status: 1},
	}
	if err := db.Create(&users).Error; err != nil {
		t.Fatalf("seed users error = %v", err)
	}

	identities := []system.UserIdentity{
		{Model: gorm.Model{ID: 11}, UserID: 1, Provider: "feishu", ProviderTenant: "tenant-a", ProviderSubject: "subject-a"},
		{Model: gorm.Model{ID: 12}, UserID: 2, Provider: "feishu", ProviderTenant: "tenant-a", ProviderSubject: "subject-b"},
	}
	if err := db.Create(&identities).Error; err != nil {
		t.Fatalf("seed identities error = %v", err)
	}

	passkeys := []system.UserPasskey{
		{Model: gorm.Model{ID: 21}, UserID: 1, CredentialID: "cred-a", CredentialPublicKey: "pk-a"},
		{Model: gorm.Model{ID: 22}, UserID: 2, CredentialID: "cred-b", CredentialPublicKey: "pk-b"},
	}
	if err := db.Create(&passkeys).Error; err != nil {
		t.Fatalf("seed passkeys error = %v", err)
	}

	roleUsers := []roleModel.User{
		{UserId: 1, RoleId: 1},
		{UserId: 2, RoleId: 1},
	}
	if err := db.Create(&roleUsers).Error; err != nil {
		t.Fatalf("seed role users error = %v", err)
	}

	svc := &userService{
		db:           db,
		userRepo:     repo.NewUserRepo(db, nil, nil),
		identityRepo: repo.NewUserIdentityRepo(db, nil, nil),
		passkeyRepo:  repo.NewUserPasskeyRepo(db, nil, nil),
		roleUserRepo: repo.NewRoleUserRepo(db, nil, nil),
		authRepo: &mockAuthRepo{
			HasRoleFunc: func(ctx context.Context, userID uint, role string) (bool, error) {
				return false, nil
			},
		},
	}

	errCode, err := svc.Delete(context.Background(), 1)
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if errCode != e.SUCCESS {
		t.Fatalf("Delete() errCode = %d, want %d", errCode, e.SUCCESS)
	}

	var activeUserCount int64
	if err := db.Model(&system.User{}).Where("id = ?", 1).Count(&activeUserCount).Error; err != nil {
		t.Fatalf("count active user error = %v", err)
	}
	if activeUserCount != 0 {
		t.Fatalf("active user count = %d, want 0", activeUserCount)
	}

	var deletedUser system.User
	if err := db.Unscoped().First(&deletedUser, 1).Error; err != nil {
		t.Fatalf("Unscoped().First() error = %v", err)
	}
	if deletedUser.DeletedAt.Valid == false {
		t.Fatalf("deleted user should be soft deleted")
	}

	var identityCount int64
	if err := db.Unscoped().Model(&system.UserIdentity{}).Where("user_id = ?", 1).Count(&identityCount).Error; err != nil {
		t.Fatalf("count identities error = %v", err)
	}
	if identityCount != 0 {
		t.Fatalf("identity count = %d, want 0", identityCount)
	}

	var passkeyCount int64
	if err := db.Unscoped().Model(&system.UserPasskey{}).Where("user_id = ?", 1).Count(&passkeyCount).Error; err != nil {
		t.Fatalf("count passkeys error = %v", err)
	}
	if passkeyCount != 0 {
		t.Fatalf("passkey count = %d, want 0", passkeyCount)
	}

	var roleUserCount int64
	if err := db.Model(&roleModel.User{}).Where("user_id = ?", 1).Count(&roleUserCount).Error; err != nil {
		t.Fatalf("count role users error = %v", err)
	}
	if roleUserCount != 0 {
		t.Fatalf("role user count = %d, want 0", roleUserCount)
	}

	if err := db.Unscoped().Model(&system.UserIdentity{}).Where("user_id = ?", 2).Count(&identityCount).Error; err != nil {
		t.Fatalf("count remaining identities error = %v", err)
	}
	if identityCount != 1 {
		t.Fatalf("remaining identity count = %d, want 1", identityCount)
	}

	if err := db.Unscoped().Model(&system.UserPasskey{}).Where("user_id = ?", 2).Count(&passkeyCount).Error; err != nil {
		t.Fatalf("count remaining passkeys error = %v", err)
	}
	if passkeyCount != 1 {
		t.Fatalf("remaining passkey count = %d, want 1", passkeyCount)
	}
}
