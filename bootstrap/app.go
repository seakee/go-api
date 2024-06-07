package bootstrap

import (
	"github.com/gin-gonic/gin"
	"github.com/seakee/go-api/app"
	"github.com/seakee/go-api/app/http/middleware"
	"github.com/seakee/go-api/app/pkg/trace"
	"github.com/sk-pkg/feishu"
	"github.com/sk-pkg/i18n"
	"github.com/sk-pkg/kafka"
	"github.com/sk-pkg/logger"
	"github.com/sk-pkg/mysql"
	"github.com/sk-pkg/redis"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"time"
)

type App struct {
	Config        *app.Config
	Logger        *zap.Logger
	Redis         map[string]*redis.Manager
	I18n          *i18n.Manager
	MysqlDB       map[string]*gorm.DB
	Middleware    middleware.Middleware
	KafkaProducer *kafka.Manager
	KafkaConsumer *kafka.Manager
	Mux           *gin.Engine
	Feishu        *feishu.Manager
	TraceID       *trace.ID
}

func NewApp(config *app.Config) (*App, error) {
	a := &App{Config: config, MysqlDB: map[string]*gorm.DB{}, Redis: map[string]*redis.Manager{}}

	err := a.loadLogger()
	if err != nil {
		return nil, err
	}

	a.loadRedis()

	err = a.loadFeishu()
	if err != nil {
		return nil, err
	}

	err = a.loadI18n()
	if err != nil {
		return nil, err
	}

	err = a.loadDB()
	if err != nil {
		return nil, err
	}

	a.loadHTTPMiddlewares()
	a.loadMux()

	err = a.loadKafka()
	if err != nil {
		return nil, err
	}

	return a, nil
}

// Start 启动应用
func (a *App) Start() {
	// 启动HTTP服务
	go a.startHTTPServer()
	// 启动kafka消费
	go a.startKafkaConsumer()
	// 启动调度任务
	go a.startSchedule()
}

// loadTrace 加载 TraceID
func (a *App) loadTrace() {
	a.TraceID = trace.NewTraceID()
}

// loadLogger 加载日志模块
func (a *App) loadLogger() error {
	var err error
	a.Logger, err = logger.New(
		logger.WithLevel(a.Config.Log.Level),
		logger.WithDriver(a.Config.Log.Driver),
		logger.WithLogPath(a.Config.Log.LogPath),
	)

	if err == nil {
		a.Logger.Info("Loggers loaded successfully")
	}

	return err
}

// loadRedis 加载Redis模块
func (a *App) loadRedis() {
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

	a.Logger.Info("Redis loaded successfully")
}

// loadI18n 加载国际化模块
func (a *App) loadI18n() error {
	var err error
	a.I18n, err = i18n.New(
		i18n.WithDebugMode(a.Config.System.DebugMode),
		i18n.WithEnvKey(a.Config.System.EnvKey),
		i18n.WithDefaultLang(a.Config.System.DefaultLang),
		i18n.WithLangDir(a.Config.System.LangDir),
	)

	if err == nil {
		a.Logger.Info("I18n loaded successfully")
	}

	return err
}

// loadDB 加载数据库模块
func (a *App) loadDB() error {

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
				// TODO mongo初始化逻辑
			}
		}
	}

	a.Logger.Info("Databases loaded successfully")

	return nil
}

// loadFeishu 加载飞书模块
func (a *App) loadFeishu() error {
	var err error

	if a.Config.Feishu.Enable {
		a.Feishu, err = feishu.New(
			feishu.WithGroupWebhook(a.Config.Feishu.GroupWebhook),
			feishu.WithAppID(a.Config.Feishu.AppID),
			feishu.WithAppSecret(a.Config.Feishu.AppSecret),
			feishu.WithEncryptKey(a.Config.Feishu.EncryptKey),
			feishu.WithRedis(a.Redis["go-api"]),
			feishu.WithLog(a.Logger),
		)

		if err == nil {
			a.Logger.Info("Feishu loaded successfully")
		}
	}

	return err
}
