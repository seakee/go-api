package bootstrap

import (
	"github.com/seakee/go-api/app/job"
	"github.com/seakee/go-api/app/pkg/schedule"
)

func (a *App) startSchedule() {
	s := schedule.New(a.Logger, a.Redis["go-api"])

	job.Register(a.Logger, a.Redis, a.MysqlDB, a.Feishu, s)

	s.Start()

	a.Logger.Info("Schedule loaded successfully")
}
