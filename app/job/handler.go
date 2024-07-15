// Copyright 2024 Seakee.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package job

import (
	"github.com/seakee/go-api/app/job/monitor"
	"github.com/seakee/go-api/app/pkg/schedule"
	"github.com/sk-pkg/feishu"
	"github.com/sk-pkg/logger"
	"github.com/sk-pkg/redis"
	"gorm.io/gorm"
)

func Register(logger *logger.Manager, redis map[string]*redis.Manager, db map[string]*gorm.DB, feishu *feishu.Manager, s *schedule.Schedule) {
	// Monitor broadband public network IP changes
	ipMonitor := monitor.NewIpMonitor(logger, redis["go-api"])
	s.AddJob("IpMonitor", ipMonitor).PerMinuit(5).WithoutOverlapping()
}
