package bootstrap

import (
	"context"
	"github.com/seakee/go-api/app/job"
	"github.com/seakee/go-api/app/pkg/schedule"
)

func (a *App) startSchedule(ctx context.Context) {
	s := schedule.New(a.Logger, a.Redis["go-api"], a.TraceID)

	job.Register(a.Logger, a.Redis, a.MysqlDB, a.Feishu, s)

	s.Start()

	a.Logger.Info(ctx, "Schedule loaded successfully")
}
