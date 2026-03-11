package system

import (
	"context"
	"testing"

	"github.com/glebarez/sqlite"
	model "github.com/seakee/go-api/app/model/system"
	"gorm.io/gorm"
)

func TestUserRepo_GetOAuthUser(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open() error = %v", err)
	}

	if err := db.AutoMigrate(&model.User{}); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}

	users := []model.User{
		{Model: gorm.Model{ID: 1}, UserName: "super-admin", FeishuId: "admin-feishu"},
		{Model: gorm.Model{ID: 2}, UserName: "normal-user", FeishuId: "user-feishu", WechatId: "wechat-user"},
	}
	if err := db.Create(&users).Error; err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	repo := userRepo{db: db}

	tests := []struct {
		name      string
		oauthType string
		id        string
		wantID    uint
		wantNil   bool
	}{
		{
			name:      "empty oauth id does not fall back to first user",
			oauthType: "feishu",
			id:        "",
			wantNil:   true,
		},
		{
			name:      "trimmed empty oauth id does not fall back to first user",
			oauthType: "feishu",
			id:        "   ",
			wantNil:   true,
		},
		{
			name:      "feishu id matches explicit user",
			oauthType: "feishu",
			id:        "user-feishu",
			wantID:    2,
		},
		{
			name:      "wechat id matches explicit user",
			oauthType: "wechat",
			id:        "wechat-user",
			wantID:    2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := repo.GetOAuthUser(context.Background(), tt.oauthType, tt.id)
			if err != nil {
				t.Fatalf("GetOAuthUser() error = %v", err)
			}
			if tt.wantNil {
				if user != nil {
					t.Fatalf("GetOAuthUser() user = %+v, want nil", user)
				}
				return
			}
			if user == nil {
				t.Fatalf("GetOAuthUser() user = nil, want id %d", tt.wantID)
			}
			if user.ID != tt.wantID {
				t.Fatalf("GetOAuthUser() id = %d, want %d", user.ID, tt.wantID)
			}
		})
	}
}
