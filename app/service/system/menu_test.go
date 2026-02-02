package system

import (
	"context"
	"errors"
	"testing"

	"gorm.io/gorm"

	"github.com/seakee/go-api/app/model/system"
	"github.com/seakee/go-api/app/pkg/e"
)

// TestMenuService_List tests fetching the menu list.
func TestMenuService_List(t *testing.T) {
	tests := []struct {
		name      string
		mockMenus system.MenuList
		mockErr   error
		wantLen   int
		wantErr   bool
	}{
		{
			name: "get menu list successfully",
			mockMenus: system.MenuList{
				{Model: gorm.Model{ID: 1}, Name: "Dashboard", Path: "/dashboard"},
				{Model: gorm.Model{ID: 2}, Name: "Users", Path: "/users"},
			},
			mockErr: nil,
			wantLen: 2,
			wantErr: false,
		},
		{
			name:      "empty menu list",
			mockMenus: system.MenuList{},
			mockErr:   nil,
			wantLen:   0,
			wantErr:   false,
		},
		{
			name:      "query error",
			mockMenus: nil,
			mockErr:   errors.New("database error"),
			wantLen:   0,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockMenuRepo := &mockMenuRepo{
				ListFunc: func(ctx context.Context) (system.MenuList, error) {
					return tt.mockMenus, tt.mockErr
				},
			}

			svc := &menuService{
				menuRepo: mockMenuRepo,
			}

			list, err := svc.List(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("List() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(list) != tt.wantLen {
				t.Errorf("List() len = %v, want %v", len(list), tt.wantLen)
			}
		})
	}
}

// TestMenuService_Detail tests fetching menu detail.
func TestMenuService_Detail(t *testing.T) {
	tests := []struct {
		name        string
		menuID      uint
		mockMenu    *system.Menu
		mockErr     error
		wantErrCode int
		wantErr     bool
	}{
		{
			name:   "get menu detail successfully",
			menuID: 1,
			mockMenu: &system.Menu{
				Model: gorm.Model{ID: 1},
				Name:  "Dashboard",
				Path:  "/dashboard",
			},
			mockErr:     nil,
			wantErrCode: e.SUCCESS,
			wantErr:     false,
		},
		{
			name:        "menu not found",
			menuID:      999,
			mockMenu:    nil,
			mockErr:     nil,
			wantErrCode: e.MenuNotFound,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockMenuRepo := &mockMenuRepo{
				DetailByIDFunc: func(ctx context.Context, id uint) (*system.Menu, error) {
					if id == tt.menuID {
						return tt.mockMenu, tt.mockErr
					}
					return nil, nil
				},
			}

			svc := &menuService{
				menuRepo: mockMenuRepo,
			}

			menu, errCode, err := svc.Detail(context.Background(), tt.menuID)

			if (err != nil) != tt.wantErr {
				t.Errorf("Detail() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if errCode != tt.wantErrCode {
				t.Errorf("Detail() errCode = %v, want %v", errCode, tt.wantErrCode)
				return
			}
			if tt.mockMenu != nil && menu == nil {
				t.Errorf("Detail() menu should not be nil")
			}
		})
	}
}

// TestMenuService_Delete tests deleting a menu.
func TestMenuService_Delete(t *testing.T) {
	tests := []struct {
		name          string
		menuID        uint
		mockCount     int64
		mockCountErr  error
		mockDeleteErr error
		wantErrCode   int
		wantErr       bool
	}{
		{
			name:          "delete menu successfully",
			menuID:        1,
			mockCount:     0,
			mockCountErr:  nil,
			mockDeleteErr: nil,
			wantErrCode:   e.SUCCESS,
			wantErr:       false,
		},
		{
			name:          "menu has sub-menus",
			menuID:        1,
			mockCount:     2,
			mockCountErr:  nil,
			mockDeleteErr: nil,
			wantErrCode:   e.MenuHasSubMenu,
			wantErr:       false,
		},
		{
			name:          "delete failed",
			menuID:        1,
			mockCount:     0,
			mockCountErr:  nil,
			mockDeleteErr: errors.New("database error"),
			wantErrCode:   e.ERROR,
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockMenuRepo := &mockMenuRepo{
				CountFunc: func(ctx context.Context, menu *system.Menu) (int64, error) {
					return tt.mockCount, tt.mockCountErr
				},
				DeleteByIDFunc: func(ctx context.Context, id uint) error {
					return tt.mockDeleteErr
				},
			}

			svc := &menuService{
				menuRepo: mockMenuRepo,
			}

			errCode, err := svc.Delete(context.Background(), tt.menuID)

			if (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if errCode != tt.wantErrCode {
				t.Errorf("Delete() errCode = %v, want %v", errCode, tt.wantErrCode)
			}
		})
	}
}
