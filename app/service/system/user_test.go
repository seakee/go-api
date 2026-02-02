package system

import (
	"context"
	"errors"
	"testing"

	"gorm.io/gorm"

	"github.com/seakee/go-api/app/model/system"
	"github.com/seakee/go-api/app/pkg/e"
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
				Account:  "testuser",
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
			wantErrCode: e.AccountNotFound,
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
		name       string
		page       int
		size       int
		mockUsers  []system.User
		mockCount  int64
		mockErr    error
		wantTotal  int64
		wantLen    int
		wantErr    bool
	}{
		{
			name: "get user list successfully",
			page: 1,
			size: 10,
			mockUsers: []system.User{
				{Model: gorm.Model{ID: 1}, Account: "user1"},
				{Model: gorm.Model{ID: 2}, Account: "user2"},
			},
			mockCount: 2,
			mockErr:   nil,
			wantTotal: 2,
			wantLen:   2,
			wantErr:   false,
		},
		{
			name:       "empty list",
			page:       1,
			size:       10,
			mockUsers:  []system.User{},
			mockCount:  0,
			mockErr:    nil,
			wantTotal:  0,
			wantLen:    0,
			wantErr:    false,
		},
		{
			name:       "query error",
			page:       1,
			size:       10,
			mockUsers:  nil,
			mockCount:  0,
			mockErr:    errors.New("database error"),
			wantTotal:  0,
			wantLen:    0,
			wantErr:    true,
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
