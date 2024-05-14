package job

import (
	"github.com/seakee/go-api/app/job/monitor"
	"github.com/seakee/go-api/app/pkg/schedule"
	"github.com/sk-pkg/feishu"
	"github.com/sk-pkg/redis"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func Register(logger *zap.Logger, redis map[string]*redis.Manager, db map[string]*gorm.DB, feishu *feishu.Manager, s *schedule.Schedule) {
	// Monitor broadband public network IP changes
	ipMonitor := monitor.NewIpMonitor(logger, redis["go-api"])
	s.AddJob("IpMonitor", ipMonitor).PerMinuit(5).WithoutOverlapping()
}
