package auth

import (
	"github.com/seakee/go-api/app/model/auth"
	"github.com/sk-pkg/redis"
	"gorm.io/gorm"
)

type (
	Repo interface {
		GetApp(*auth.App) (*auth.App, error)
		Create(*auth.App) (uint, error)
		ExistAppByName(string) (bool, error)
	}

	repo struct {
		redis *redis.Manager
		db    *gorm.DB
	}
)

func (r repo) ExistAppByName(name string) (exist bool, err error) {
	app := &auth.App{AppName: name}
	a, err := app.First(r.db)
	if a != nil {
		exist = true
	}

	return
}

func (r repo) Create(app *auth.App) (uint, error) {
	return app.Create(r.db)
}

func (r repo) GetApp(app *auth.App) (*auth.App, error) {
	return app.First(r.db)
}

func NewAppRepo(db *gorm.DB, redis *redis.Manager) Repo {
	return &repo{redis: redis, db: db}
}
