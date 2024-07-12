package auth

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

type App struct {
	gorm.Model

	AppName     string `gorm:"column:app_name" json:"app_name"`
	AppID       string `gorm:"column:app_id" json:"app_id"`
	AppSecret   string `gorm:"column:app_secret" json:"app_secret"`
	RedirectUri string `gorm:"column:redirect_uri" json:"redirect_uri"`
	Description string `gorm:"column:description" json:"description"`
	Status      uint8  `gorm:"column:status" json:"status"`
}

func (a *App) TableName() string {
	return "auth_app"
}

func (a *App) First(ctx context.Context, db *gorm.DB) (*App, error) {
	var app App

	err := db.WithContext(ctx).Where(a).First(&app).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, fmt.Errorf("first failed: %w", err)
	}

	return &app, nil
}

func (a *App) Last(ctx context.Context, db *gorm.DB) (*App, error) {
	var app App

	err := db.WithContext(ctx).Where(a).Order("id desc").First(&app).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("find last failed: %w", err)
	}

	return &app, nil
}

func (a *App) Create(ctx context.Context, db *gorm.DB) (uint, error) {
	if err := db.WithContext(ctx).Create(a).Error; err != nil {
		return 0, fmt.Errorf("create failed: %w", err)
	}

	return a.ID, nil
}

func (a *App) Delete(ctx context.Context, db *gorm.DB) error {
	if err := db.WithContext(ctx).Delete(a).Error; err != nil {
		return fmt.Errorf("delete failed: %w", err)
	}

	return nil
}

func (a *App) Updates(ctx context.Context, db *gorm.DB, updates map[string]interface{}) error {
	if err := db.WithContext(ctx).Model(a).Updates(updates).Error; err != nil {
		return fmt.Errorf("updates failed: %w", err)
	}

	return nil
}

func (a *App) List(ctx context.Context, db *gorm.DB) ([]App, error) {
	var apps []App

	err := db.WithContext(ctx).Where(a).Find(&apps).Error
	if err != nil {
		return nil, fmt.Errorf("list failed: %w", err)
	}

	return apps, nil
}

func (a *App) ListByArgs(ctx context.Context, db *gorm.DB, query interface{}, args ...interface{}) ([]App, error) {
	var apps []App

	err := db.WithContext(ctx).Where(query, args...).Order("id desc").Find(&apps).Error
	if err != nil {
		return nil, fmt.Errorf("list by args failed: %w", err)
	}

	return apps, nil
}

func (a *App) CountByArgs(ctx context.Context, db *gorm.DB, query interface{}, args ...interface{}) (int64, error) {
	var count int64

	err := db.WithContext(ctx).Where(query, args...).Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("count by args failed: %w", err)
	}

	return count, nil
}

func (a *App) Count(ctx context.Context, db *gorm.DB) (int64, error) {
	var count int64

	err := db.WithContext(ctx).Where(a).Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("count failed: %w", err)
	}

	return count, nil
}

func (a *App) BatchInsert(ctx context.Context, db *gorm.DB, apps []App) error {
	if err := db.WithContext(ctx).Create(&apps).Error; err != nil {
		return fmt.Errorf("batch insert failed: %w", err)
	}

	return nil
}

func (a *App) FindWithPagination(ctx context.Context, db *gorm.DB, page, size int) ([]App, error) {
	var apps []App

	err := db.WithContext(ctx).Where(a).Offset((page - 1) * size).Limit(size).Find(&apps).Error
	if err != nil {
		return nil, fmt.Errorf("find with pagination failed: %w", err)
	}

	return apps, nil
}

func (a *App) FindWithSort(ctx context.Context, db *gorm.DB, sort string) ([]App, error) {
	var apps []App

	err := db.WithContext(ctx).Where(a).Order(sort).Find(&apps).Error
	if err != nil {
		return nil, fmt.Errorf("find with sort failed: %w", err)
	}

	return apps, nil
}
