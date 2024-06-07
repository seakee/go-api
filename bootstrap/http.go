package bootstrap

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/seakee/go-api/app/http/middleware"
	"github.com/seakee/go-api/app/http/router"
	"github.com/sk-pkg/monitor"
	"go.uber.org/zap"
	"net/http"
	"os"
	"time"
)

// startHTTPServer 启动HTTP服务
func (a *App) startHTTPServer() {
	gin.SetMode(a.Config.System.RunMode)

	core := &router.Core{
		Logger:        a.Logger,
		Redis:         a.Redis,
		I18n:          a.I18n,
		MysqlDB:       a.MysqlDB,
		Middleware:    a.Middleware,
		KafkaProducer: a.KafkaProducer,
	}

	serverHandler := router.New(a.Mux, core)

	readTimeout := a.Config.System.ReadTimeout * time.Second
	writeTimeout := a.Config.System.WriteTimeout * time.Second
	maxHeaderBytes := 1 << 20

	server := &http.Server{
		Addr:           a.Config.System.HTTPPort,
		Handler:        serverHandler,
		ReadTimeout:    readTimeout,
		WriteTimeout:   writeTimeout,
		MaxHeaderBytes: maxHeaderBytes,
	}

	// 监听HTTP服务
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		a.Logger.Fatal("http server startup err", zap.Error(err))
	}
}

// loadMux 加载gin引擎
func (a *App) loadMux() {
	mux := gin.New()

	mux.Use(a.Middleware.SetTraceID())

	if a.Config.System.DebugMode {
		mux.Use(a.Middleware.RequestLogger())
	}

	mux.Use(a.Middleware.Cors())
	mux.Use(gin.Recovery())

	a.loadPanicRobot(mux) // panic监控

	a.Mux = mux

	a.Logger.Info("Mux loaded successfully")
}

// loadPanicRobot 加载panic监控机器人
func (a *App) loadPanicRobot(mux *gin.Engine) {
	panicRobot, err := monitor.NewPanicRobot(
		monitor.PanicRobotEnable(a.Config.Monitor.PanicRobot.Enable),
		monitor.PanicRobotEnv(os.Getenv(a.Config.System.EnvKey)),
		monitor.PanicRobotWechatEnable(a.Config.Monitor.PanicRobot.Wechat.Enable),
		monitor.PanicRobotWechatPushUrl(a.Config.Monitor.PanicRobot.Wechat.PushUrl),
		monitor.PanicRobotFeishuEnable(a.Config.Monitor.PanicRobot.Feishu.Enable),
		monitor.PanicRobotFeishuPushUrl(a.Config.Monitor.PanicRobot.Feishu.PushUrl),
	)

	if err == nil {
		mux.Use(panicRobot.Middleware())
	}
}

// loadHTTPMiddlewares 加载HTTP中间件
func (a *App) loadHTTPMiddlewares() {
	a.Middleware = middleware.New(a.Logger, a.I18n, a.MysqlDB, a.Redis, a.TraceID)
	a.Logger.Info("Middlewares loaded successfully")
}
