package monitor

import (
	"context"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/seakee/go-api/app/pkg/schedule"
	"github.com/sk-pkg/feishu"
	"github.com/sk-pkg/logger"
	"github.com/sk-pkg/redis"
	"go.uber.org/zap"
	"strings"
)

const (
	CheckCNIpApi = "http://members.3322.org/dyndns/getip"
	CheckIpApi   = "http://whatismyip.akamai.com/"
	lastIpKey    = "monitor:ip:lastIp"
)

type ipHandler struct {
	done   chan struct{}
	error  chan error
	logger *logger.Manager
	redis  *redis.Manager
	lastIp string
	feishu *feishu.Manager
}

func (ih *ipHandler) setLastIp() {
	lastIp, err := ih.redis.GetString(lastIpKey)
	if err != nil {
		ih.error <- fmt.Errorf("failed to get last IP from Redis: %w", err)
		return
	}

	ih.lastIp = lastIp
}

func (ih *ipHandler) Exec(ctx context.Context) {
	ih.setLastIp()

	client := resty.New()
	res, err := client.R().Get(CheckCNIpApi)
	if err == nil && res != nil && res.StatusCode() == 200 {
		currentIp := strings.TrimRight(string(res.Body()), "\n")
		if ih.lastIp != currentIp && currentIp != "" {
			ih.logger.Info(ctx, "IP has changed", zap.String("last ip", ih.lastIp), zap.String("current ip", currentIp))
			ih.lastIp = currentIp

			if err = ih.redis.SetString(lastIpKey, currentIp, 0); err != nil {
				ih.error <- fmt.Errorf("failed to set last IP (%s) in Redis: %w", currentIp, err)
			}
		}
	} else if err != nil {
		ih.error <- fmt.Errorf("failed to check IP from %s: %w", CheckCNIpApi, err)
	}

	ih.done <- struct{}{}
}

func (ih *ipHandler) Error() <-chan error {
	return ih.error
}

func (ih *ipHandler) Done() <-chan struct{} {
	return ih.done
}

func NewIpMonitor(logger *logger.Manager, redis *redis.Manager) schedule.HandlerFunc {
	return &ipHandler{
		done:   make(chan struct{}),
		error:  make(chan error),
		logger: logger,
		lastIp: "",
		redis:  redis,
	}
}
