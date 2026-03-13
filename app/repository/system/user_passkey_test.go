package system

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	model "github.com/seakee/go-api/app/model/system"
	"gorm.io/gorm"
)

func TestUserPasskeyRepo_CRUDAndCount(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open() error = %v", err)
	}

	if err = db.AutoMigrate(&model.User{}, &model.UserPasskey{}); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}

	users := []model.User{
		{Model: gorm.Model{ID: 1}, UserName: "u1"},
		{Model: gorm.Model{ID: 2}, UserName: "u2"},
	}
	if err = db.Create(&users).Error; err != nil {
		t.Fatalf("Create users error = %v", err)
	}

	repo := NewUserPasskeyRepo(db, nil, nil)
	ctx := context.Background()

	created1, err := repo.Create(ctx, &model.UserPasskey{
		UserID:              1,
		CredentialID:        "cred-1",
		CredentialPublicKey: "pk-1",
		SignCount:           1,
		DisplayName:         "device-1",
	})
	if err != nil {
		t.Fatalf("Create() credential 1 error = %v", err)
	}

	created2, err := repo.Create(ctx, &model.UserPasskey{
		UserID:              1,
		CredentialID:        "cred-2",
		CredentialPublicKey: "pk-2",
		SignCount:           2,
		DisplayName:         "device-2",
	})
	if err != nil {
		t.Fatalf("Create() credential 2 error = %v", err)
	}

	created3, err := repo.Create(ctx, &model.UserPasskey{
		UserID:              2,
		CredentialID:        "cred-3",
		CredentialPublicKey: "pk-3",
		SignCount:           3,
		DisplayName:         "device-3",
	})
	if err != nil {
		t.Fatalf("Create() credential 3 error = %v", err)
	}

	detail, err := repo.DetailByCredentialID(ctx, "cred-2")
	if err != nil {
		t.Fatalf("DetailByCredentialID() error = %v", err)
	}
	if detail == nil || detail.ID != created2.ID {
		t.Fatalf("DetailByCredentialID() = %+v, want id %d", detail, created2.ID)
	}

	list, err := repo.ListByUserID(ctx, 1)
	if err != nil {
		t.Fatalf("ListByUserID() error = %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("ListByUserID() len = %d, want 2", len(list))
	}
	if list[0].ID != created1.ID || list[1].ID != created2.ID {
		t.Fatalf("ListByUserID() order = [%d, %d], want [%d, %d]", list[0].ID, list[1].ID, created1.ID, created2.ID)
	}

	countUser1, err := repo.CountByUserID(ctx, 1)
	if err != nil {
		t.Fatalf("CountByUserID() error = %v", err)
	}
	if countUser1 != 2 {
		t.Fatalf("CountByUserID() = %d, want 2", countUser1)
	}

	counts, err := repo.CountByUserIDs(ctx, []uint{1, 2, 99})
	if err != nil {
		t.Fatalf("CountByUserIDs() error = %v", err)
	}
	if counts[1] != 2 || counts[2] != 1 || counts[99] != 0 {
		t.Fatalf("CountByUserIDs() = %+v, want map[1:2 2:1 99:0]", counts)
	}

	usedAt := time.Unix(1700000000, 0).UTC()
	if err = repo.UpdateSignCount(ctx, created1.ID, 88); err != nil {
		t.Fatalf("UpdateSignCount() error = %v", err)
	}
	if err = repo.UpdateLastUsedAt(ctx, created1.ID, usedAt); err != nil {
		t.Fatalf("UpdateLastUsedAt() error = %v", err)
	}

	updated, err := repo.DetailByCredentialID(ctx, "cred-1")
	if err != nil {
		t.Fatalf("DetailByCredentialID() after update error = %v", err)
	}
	if updated.SignCount != 88 {
		t.Fatalf("updated.SignCount = %d, want 88", updated.SignCount)
	}
	if updated.LastUsedAt == nil || !updated.LastUsedAt.UTC().Equal(usedAt) {
		t.Fatalf("updated.LastUsedAt = %v, want %v", updated.LastUsedAt, usedAt)
	}

	if err = repo.DeleteByIDAndUserID(ctx, created2.ID, 1); err != nil {
		t.Fatalf("DeleteByIDAndUserID() error = %v", err)
	}
	if err = repo.DeleteByIDAndUserID(ctx, created2.ID, 1); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("DeleteByIDAndUserID() error = %v, want ErrRecordNotFound", err)
	}

	if err = repo.DeleteByUserID(ctx, 1); err != nil {
		t.Fatalf("DeleteByUserID() error = %v", err)
	}

	countUser1, err = repo.CountByUserID(ctx, 1)
	if err != nil {
		t.Fatalf("CountByUserID() after delete error = %v", err)
	}
	if countUser1 != 0 {
		t.Fatalf("CountByUserID() after delete = %d, want 0", countUser1)
	}

	remaining, err := repo.DetailByIDAndUserID(ctx, created3.ID, 2)
	if err != nil {
		t.Fatalf("DetailByIDAndUserID() for user2 error = %v", err)
	}
	if remaining == nil {
		t.Fatalf("DetailByIDAndUserID() for user2 = nil, want record")
	}
}

func TestUserPasskeyRepo_ValidateRequiredIDs(t *testing.T) {
	repo := userPasskeyRepo{}
	ctx := context.Background()

	if err := repo.UpdateSignCount(ctx, 0, 1); err == nil {
		t.Fatalf("UpdateSignCount() error = nil, want non-nil")
	}
	if err := repo.UpdateLastUsedAt(ctx, 0, time.Now()); err == nil {
		t.Fatalf("UpdateLastUsedAt() error = nil, want non-nil")
	}
	if err := repo.DeleteByIDAndUserID(ctx, 0, 1); err == nil {
		t.Fatalf("DeleteByIDAndUserID() with empty passkey id error = nil, want non-nil")
	}
	if err := repo.DeleteByIDAndUserID(ctx, 1, 0); err == nil {
		t.Fatalf("DeleteByIDAndUserID() with empty user id error = nil, want non-nil")
	}
	if err := repo.DeleteByID(ctx, 0); err == nil {
		t.Fatalf("DeleteByID() error = nil, want non-nil")
	}
	if err := repo.DeleteByUserID(ctx, 0); err == nil {
		t.Fatalf("DeleteByUserID() error = nil, want non-nil")
	}
}
