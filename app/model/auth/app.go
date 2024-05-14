package auth

import (
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

func (a *App) First(db *gorm.DB) (app *App, err error) {
	err = db.Where(a).First(&app).Error

	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	return app, err
}

func (a *App) Create(db *gorm.DB) (id uint, err error) {
	if err = db.Create(a).Error; err != nil {
		return 0, fmt.Errorf("create failed: %w", err)
	}

	id = a.ID

	return
}

func (a *App) Delete(db *gorm.DB) (err error) {
	if err = db.Delete(a).Error; err != nil {
		return fmt.Errorf("delete failed: %w", err)
	}
	return
}

func (a *App) Updates(db *gorm.DB, m map[string]interface{}) (err error) {
	if err = db.Model(&App{}).Where("id = ?", a.ID).Updates(m).Error; err != nil {
		return fmt.Errorf("updates failed: %w", err)
	}
	return
}
