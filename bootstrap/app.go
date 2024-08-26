// Copyright 2024 Seakee.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

// Package bootstrap provides functionality to initialize and start the application.
// It handles the setup of various components such as logging, databases, caching,
// message queues, and HTTP servers.
package bootstrap

import (
	"context"
	"github.com/sk-pkg/notify/lark"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/qiniu/qmgo"
	"github.com/qiniu/qmgo/options"
	"github.com/seakee/go-api/app"
	"github.com/seakee/go-api/app/http/middleware"
	"github.com/seakee/go-api/app/pkg/trace"
	"github.com/sk-pkg/i18n"
	"github.com/sk-pkg/kafka"
	"github.com/sk-pkg/logger"
	"github.com/sk-pkg/mysql"
	"github.com/sk-pkg/notify"
	"github.com/sk-pkg/redis"
	mgOpt "go.mongodb.org/mongo-driver/mongo/options"
	"gorm.io/gorm"
)

// App represents the main application structure, containing all necessary components and configurations.
type App struct {
	Config        *app.Config
	Logger        *logger.Manager
	Redis         map[string]*redis.Manager
	I18n          *i18n.Manager
	MysqlDB       map[string]*gorm.DB
	MongoDB       map[string]*qmgo.Database
	Middleware    middleware.Middleware
	KafkaProducer *kafka.Manager
	KafkaConsumer *kafka.Manager
	Mux           *gin.Engine
	Notify        *notify.Manager
	TraceID       *trace.ID
}

// NewApp creates and initializes a new App instance.
//
// Parameters:
//   - config: A pointer to the application configuration.
//
// Returns:
//   - *App: A pointer to the initialized App instance.
//   - error: An error if any initialization step fails.
func NewApp(config *app.Config) (*App, error) {
	a := &App{
		Config:  config,
		MysqlDB: map[string]*gorm.DB{},
		MongoDB: map[string]*qmgo.Database{},
		Redis:   map[string]*redis.Manager{},
	}

	// Initialize components
	a.loadTrace()

	ctx := context.WithValue(context.Background(), logger.TraceIDKey, a.TraceID.New())

	err := a.loadLogger(ctx)
	if err != nil {
		return nil, err
	}

	err = a.loadRedis(ctx)
	if err != nil {
		return nil, err
	}

	err = a.loadNotify()
	if err != nil {
		return a, err
	}

	err = a.loadI18n(ctx)
	if err != nil {
		return nil, err
	}

	err = a.loadDB(ctx)
	if err != nil {
		return nil, err
	}

	a.loadHTTPMiddlewares(ctx)
	a.loadMux(ctx)

	err = a.loadKafka(ctx)
	if err != nil {
		return nil, err
	}

	return a, nil
}

// Start initiates the application by starting the HTTP server, Kafka consumer, and scheduled tasks.
// It runs each component in a separate goroutine.
func (a *App) Start() {
	ctx := context.WithValue(context.Background(), logger.TraceIDKey, a.TraceID.New())
	// Start HTTP server
	go a.startHTTPServer(ctx)
	// Start Kafka consumer
	go a.startKafkaConsumer(ctx)
	// Start scheduled tasks
	go a.startSchedule(ctx)
}

// loadTrace initializes the TraceID component.
func (a *App) loadTrace() {
	a.TraceID = trace.NewTraceID()
}

// loadLogger initializes the logging component.
//
// Parameters:
//   - ctx: The context for the operation.
//
// Returns:
//   - error: An error if the logger initialization fails.
func (a *App) loadLogger(ctx context.Context) error {
	var err error
	a.Logger, err = logger.New(
		logger.WithLevel(a.Config.Log.Level),
		logger.WithDriver(a.Config.Log.Driver),
		logger.WithLogPath(a.Config.Log.LogPath),
	)

	if err == nil {
		a.Logger.Info(ctx, "Loggers loaded successfully")
	}

	return err
}

// loadRedis initializes the Redis caching component.
//
// Parameters:
//   - ctx: The context for the operation.
func (a *App) loadRedis(ctx context.Context) error {
	for _, cfg := range a.Config.Redis {
		if cfg.Enable {
			r, err := redis.New(
				redis.WithPrefix(cfg.Prefix),
				redis.WithAddress(cfg.Host),
				redis.WithPassword(cfg.Auth),
				redis.WithIdleTimeout(cfg.IdleTimeout*time.Minute),
				redis.WithMaxActive(cfg.MaxActive),
				redis.WithMaxIdle(cfg.MaxIdle),
				redis.WithDB(cfg.DB),
			)
			if err != nil {
				return err
			}

			a.Redis[cfg.Name] = r
		}
	}

	a.Logger.Info(ctx, "Redis loaded successfully")

	return nil
}

// loadI18n initializes the internationalization component.
//
// Parameters:
//   - ctx: The context for the operation.
//
// Returns:
//   - error: An error if the i18n initialization fails.
func (a *App) loadI18n(ctx context.Context) error {
	var err error
	a.I18n, err = i18n.New(
		i18n.WithDebugMode(a.Config.System.DebugMode),
		i18n.WithEnvKey(a.Config.System.EnvKey),
		i18n.WithDefaultLang(a.Config.System.DefaultLang),
		i18n.WithLangDir(a.Config.System.LangDir),
	)

	if err == nil {
		a.Logger.Info(ctx, "I18n loaded successfully")
	}

	return err
}

// loadDB initializes the database components (MySQL and MongoDB).
//
// Parameters:
//   - ctx: The context for the operation.
//
// Returns:
//   - error: An error if any database initialization fails.
func (a *App) loadDB(ctx context.Context) error {
	for _, db := range a.Config.Databases {
		if db.Enable {
			switch db.DbType {
			case "mysql":
				d, err := mysql.New(mysql.WithConfigs(
					mysql.Config{
						User:     db.DbUsername,
						Password: db.DbPassword,
						Host:     db.DbHost,
						DBName:   db.DbName,
					}),
					mysql.WithConnMaxLifetime(db.DbMaxLifetime*time.Hour),
					mysql.WithMaxIdleConn(db.DbMaxIdleConn),
					mysql.WithMaxOpenConn(db.DbMaxOpenConn),
				)

				if err != nil {
					return err
				}

				a.MysqlDB[db.DbName] = d
			case "mongo":
				maxPoolSize := uint64(db.DbMaxOpenConn)
				minPoolSize := uint64(db.DbMaxIdleConn)
				maxConnIdleTime := db.DbMaxLifetime * time.Hour

				opts := options.ClientOptions{ClientOptions: &mgOpt.ClientOptions{MaxConnIdleTime: &maxConnIdleTime}}
				cli, err := qmgo.NewClient(ctx, &qmgo.Config{
					Uri:         db.DbHost,
					MaxPoolSize: &maxPoolSize,
					MinPoolSize: &minPoolSize,
					Auth: &qmgo.Credential{
						AuthMechanism: db.AuthMechanism,
						AuthSource:    db.DbName,
						Username:      db.DbUsername,
						Password:      db.DbPassword,
					},
				}, opts)
				if err != nil {
					return err
				}

				a.MongoDB[db.DbName] = cli.Database(db.DbName)
			}
		}
	}

	a.Logger.Info(ctx, "Databases loaded successfully")

	return nil
}

// loadNotify initializes the notification component.
func (a *App) loadNotify() error {
	larksCount := len(a.Config.Notify.Lark.Larks)
	larks := make(map[string]lark.Lark, larksCount)
	if larksCount > 0 {
		for name, l := range a.Config.Notify.Lark.Larks {
			larks[name] = lark.Lark{
				AppType:   l.AppType,
				AppID:     l.AppID,
				AppSecret: l.AppSecret,
			}
		}
	}

	manager, err := notify.New(
		notify.OptDefaultChannel(notify.Channel(a.Config.Notify.DefaultChannel)),
		notify.OptDefaultLevel(notify.Level(a.Config.Notify.DefaultLevel)),
		notify.OptLarkConfig(lark.Config{
			Enabled:                a.Config.Notify.Lark.Enable,
			DefaultSendChannelName: a.Config.Notify.Lark.DefaultSendChannelName,
			BotWebhooks:            a.Config.Notify.Lark.BotWebhooks,
			Larks:                  larks,
		}),
	)
	if err != nil {
		return err
	}

	a.Notify = manager

	return nil
}
