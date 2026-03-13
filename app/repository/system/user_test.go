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

	if err := db.AutoMigrate(&model.User{}, &model.UserIdentity{}); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}

	users := []model.User{
		{Model: gorm.Model{ID: 1}, UserName: "super-admin"},
		{Model: gorm.Model{ID: 2}, UserName: "normal-user"},
	}
	if err := db.Create(&users).Error; err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	identities := []model.UserIdentity{
		{UserID: 1, Provider: "feishu", ProviderTenant: "tenant-a", ProviderSubject: "admin-feishu"},
		{UserID: 2, Provider: "feishu", ProviderTenant: "tenant-a", ProviderSubject: "user-feishu"},
		{UserID: 2, Provider: "wechat", ProviderTenant: "corp-a", ProviderSubject: "wechat-user"},
	}
	if err := db.Create(&identities).Error; err != nil {
		t.Fatalf("Create identities error = %v", err)
	}

	repo := userRepo{db: db}

	tests := []struct {
		name      string
		oauthType string
		tenant    string
		id        string
		wantID    uint
		wantNil   bool
	}{
		{
			name:      "empty oauth id does not fall back to first user",
			oauthType: "feishu",
			tenant:    "tenant-a",
			id:        "",
			wantNil:   true,
		},
		{
			name:      "trimmed empty oauth id does not fall back to first user",
			oauthType: "feishu",
			tenant:    "tenant-a",
			id:        "   ",
			wantNil:   true,
		},
		{
			name:      "feishu id matches explicit user",
			oauthType: "feishu",
			tenant:    "tenant-a",
			id:        "user-feishu",
			wantID:    2,
		},
		{
			name:      "wechat id matches explicit user",
			oauthType: "wechat",
			tenant:    "corp-a",
			id:        "wechat-user",
			wantID:    2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := repo.GetOAuthUser(context.Background(), tt.oauthType, tt.tenant, tt.id)
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

func TestUserIdentityRepo_DeleteByIDAndUserID_HardDeleteAllowsRebind(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open() error = %v", err)
	}

	if err := db.AutoMigrate(&model.User{}, &model.UserIdentity{}); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}
	if err := db.Exec("CREATE UNIQUE INDEX uk_sys_user_identity_provider_subject ON sys_user_identity (provider, provider_tenant, provider_subject)").Error; err != nil {
		t.Fatalf("create unique index error = %v", err)
	}

	user := model.User{Model: gorm.Model{ID: 1}, UserName: "admin"}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("Create user error = %v", err)
	}

	identity := model.UserIdentity{
		UserID:          1,
		Provider:        "feishu",
		ProviderTenant:  "tenant-a",
		ProviderSubject: "subject-a",
	}
	if err := db.Create(&identity).Error; err != nil {
		t.Fatalf("Create identity error = %v", err)
	}

	identityRepo := NewUserIdentityRepo(db, nil, nil)
	if err := identityRepo.DeleteByIDAndUserID(context.Background(), identity.ID, 1); err != nil {
		t.Fatalf("DeleteByIDAndUserID() error = %v", err)
	}

	rebinding := model.UserIdentity{
		UserID:          1,
		Provider:        "feishu",
		ProviderTenant:  "tenant-a",
		ProviderSubject: "subject-a",
	}
	if _, err := identityRepo.Create(context.Background(), &rebinding); err != nil {
		t.Fatalf("Create() after delete error = %v", err)
	}

	var deletedCount int64
	if err := db.Unscoped().Model(&model.UserIdentity{}).
		Where("provider = ? AND provider_tenant = ? AND provider_subject = ?", "feishu", "tenant-a", "subject-a").
		Count(&deletedCount).Error; err != nil {
		t.Fatalf("count identities error = %v", err)
	}
	if deletedCount != 1 {
		t.Fatalf("Unscoped identity count = %d, want 1", deletedCount)
	}
}

func TestUserIdentityRepo_DeleteByUserID(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open() error = %v", err)
	}

	if err := db.AutoMigrate(&model.User{}, &model.UserIdentity{}); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}

	users := []model.User{
		{Model: gorm.Model{ID: 1}, UserName: "u1"},
		{Model: gorm.Model{ID: 2}, UserName: "u2"},
	}
	if err := db.Create(&users).Error; err != nil {
		t.Fatalf("Create users error = %v", err)
	}

	identities := []model.UserIdentity{
		{UserID: 1, Provider: "feishu", ProviderTenant: "tenant-a", ProviderSubject: "subject-a"},
		{UserID: 1, Provider: "wechat", ProviderTenant: "corp-a", ProviderSubject: "subject-b"},
		{UserID: 2, Provider: "feishu", ProviderTenant: "tenant-a", ProviderSubject: "subject-c"},
	}
	if err := db.Create(&identities).Error; err != nil {
		t.Fatalf("Create identities error = %v", err)
	}

	identityRepo := NewUserIdentityRepo(db, nil, nil)
	if err := identityRepo.DeleteByUserID(context.Background(), 1); err != nil {
		t.Fatalf("DeleteByUserID() error = %v", err)
	}

	var user1Count int64
	if err := db.Unscoped().Model(&model.UserIdentity{}).Where("user_id = ?", 1).Count(&user1Count).Error; err != nil {
		t.Fatalf("count user1 identities error = %v", err)
	}
	if user1Count != 0 {
		t.Fatalf("user1 identity count = %d, want 0", user1Count)
	}

	var user2Count int64
	if err := db.Unscoped().Model(&model.UserIdentity{}).Where("user_id = ?", 2).Count(&user2Count).Error; err != nil {
		t.Fatalf("count user2 identities error = %v", err)
	}
	if user2Count != 1 {
		t.Fatalf("user2 identity count = %d, want 1", user2Count)
	}
}
