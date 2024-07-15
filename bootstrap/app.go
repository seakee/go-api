// Copyright 2024 Seakee.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package bootstrap

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/qiniu/qmgo"
	"github.com/qiniu/qmgo/options"
	"github.com/seakee/go-api/app"
	"github.com/seakee/go-api/app/http/middleware"
	"github.com/seakee/go-api/app/pkg/trace"
	"github.com/sk-pkg/feishu"
	"github.com/sk-pkg/i18n"
	"github.com/sk-pkg/kafka"
	"github.com/sk-pkg/logger"
	"github.com/sk-pkg/mysql"
	"github.com/sk-pkg/redis"
	mgOpt "go.mongodb.org/mongo-driver/mongo/options"
	"gorm.io/gorm"
)

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
	Feishu        *feishu.Manager
	TraceID       *trace.ID
}

func NewApp(config *app.Config) (*App, error) {
	a := &App{
		Config:  config,
		MysqlDB: map[string]*gorm.DB{},
		MongoDB: map[string]*qmgo.Database{},
		Redis:   map[string]*redis.Manager{},
	}

	a.loadTrace()

	ctx := context.WithValue(context.Background(), logger.TraceIDKey, a.TraceID.New())

	err := a.loadLogger(ctx)
	if err != nil {
		return nil, err
	}

	a.loadRedis(ctx)

	err = a.loadFeishu(ctx)
	if err != nil {
		return nil, err
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

// Start 启动应用
func (a *App) Start() {
	ctx := context.WithValue(context.Background(), logger.TraceIDKey, a.TraceID.New())
	// 启动HTTP服务
	go a.startHTTPServer(ctx)
	// 启动kafka消费
	go a.startKafkaConsumer(ctx)
	// 启动调度任务
	go a.startSchedule(ctx)
}

// loadTrace 加载 TraceID
func (a *App) loadTrace() {
	a.TraceID = trace.NewTraceID()
}

// loadLogger 加载日志模块
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

// loadRedis 加载Redis模块
func (a *App) loadRedis(ctx context.Context) {
	for _, cfg := range a.Config.Redis {
		if cfg.Enable {
			r := redis.New(
				redis.WithPrefix(cfg.Prefix),
				redis.WithAddress(cfg.Host),
				redis.WithPassword(cfg.Auth),
				redis.WithIdleTimeout(cfg.IdleTimeout*time.Minute),
				redis.WithMaxActive(cfg.MaxActive),
				redis.WithMaxIdle(cfg.MaxIdle),
				redis.WithDB(cfg.DB),
			)

			a.Redis[cfg.Name] = r
		}
	}

	a.Logger.Info(ctx, "Redis loaded successfully")
}

// loadI18n 加载国际化模块
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

// loadDB 加载数据库模块
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

// loadFeishu 加载飞书模块
func (a *App) loadFeishu(ctx context.Context) error {
	var err error

	if a.Config.Feishu.Enable {
		a.Feishu, err = feishu.New(
			feishu.WithGroupWebhook(a.Config.Feishu.GroupWebhook),
			feishu.WithAppID(a.Config.Feishu.AppID),
			feishu.WithAppSecret(a.Config.Feishu.AppSecret),
			feishu.WithEncryptKey(a.Config.Feishu.EncryptKey),
			feishu.WithRedis(a.Redis["go-api"]),
			feishu.WithLog(a.Logger.Zap),
		)

		if err == nil {
			a.Logger.Info(ctx, "Feishu loaded successfully")
		}
	}

	return err
}
