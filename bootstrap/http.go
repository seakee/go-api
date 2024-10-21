// Copyright 2024 Seakee.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package bootstrap

import (
	"context"
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	appHttp "github.com/seakee/go-api/app/http"
	"github.com/seakee/go-api/app/http/middleware"
	"github.com/seakee/go-api/app/http/router"
	"github.com/sk-pkg/monitor"
	"go.uber.org/zap"
)

// startHTTPServer initializes and starts the HTTP server.
//
// Parameters:
//   - ctx: The context for the operation.
//
// This function sets up the Gin engine, configures the server,
// and starts listening for incoming HTTP requests.
func (a *App) startHTTPServer(ctx context.Context) {
	gin.SetMode(a.Config.System.RunMode)

	appCtx := &appHttp.Context{
		Logger:        a.Logger,
		Redis:         a.Redis,
		I18n:          a.I18n,
		MysqlDB:       a.MysqlDB,
		MongoDB:       a.MongoDB,
		Middleware:    a.Middleware,
		KafkaProducer: a.KafkaProducer,
		Notify:        a.Notify,
		Config:        a.Config,
	}

	router.Register(a.Mux, appCtx)

	readTimeout := a.Config.System.ReadTimeout * time.Second
	writeTimeout := a.Config.System.WriteTimeout * time.Second
	maxHeaderBytes := 1 << 20

	server := &http.Server{
		Addr:           a.Config.System.HTTPPort,
		Handler:        a.Mux,
		ReadTimeout:    readTimeout,
		WriteTimeout:   writeTimeout,
		MaxHeaderBytes: maxHeaderBytes,
	}

	// Start listening for HTTP requests
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		a.Logger.Fatal(ctx, "http server startup err", zap.Error(err))
	}
}

// loadMux initializes the Gin engine and sets up middleware.
//
// Parameters:
//   - ctx: The context for the operation.
//
// This function configures the Gin engine with various middleware
// and sets up panic monitoring if enabled.
func (a *App) loadMux(ctx context.Context) {
	mux := gin.New()

	mux.Use(a.Middleware.SetTraceID())

	if a.Config.System.DebugMode {
		mux.Use(a.Middleware.RequestLogger())
	}

	mux.Use(a.Middleware.Cors())
	mux.Use(gin.Recovery())

	a.loadPanicRobot(mux) // Setup panic monitoring

	a.Mux = mux

	a.Logger.Info(ctx, "Mux loaded successfully")
}

// loadPanicRobot sets up the panic monitoring robot.
//
// Parameters:
//   - mux: The Gin engine to attach the panic robot middleware to.
//
// This function initializes the panic monitoring robot with the
// configured settings and attaches its middleware to the Gin engine.
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

// loadHTTPMiddlewares initializes the HTTP middleware.
//
// Parameters:
//   - ctx: The context for the operation.
//
// This function sets up the middleware with various components
// such as logger, i18n, databases, and Redis.
func (a *App) loadHTTPMiddlewares(ctx context.Context) {
	a.Middleware = middleware.New(a.Logger, a.I18n, a.MysqlDB, a.Redis, a.TraceID)
	a.Logger.Info(ctx, "Middlewares loaded successfully")
}
