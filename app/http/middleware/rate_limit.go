package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/seakee/go-api/app/config"
	"github.com/sk-pkg/redis"
	"github.com/sk-pkg/util"
	"go.uber.org/zap"
)

// rateLimitLuaScript 将 INCR 和 EXPIRE 合为原子操作，
// 避免 EXPIRE 单独失败导致 key 永不过期（即用户被永久限流）。
const rateLimitLuaScript = `
local count = redis.call('INCR', KEYS[1])
if count == 1 then
    redis.call('EXPIRE', KEYS[1], tonumber(ARGV[1]))
end
return count
`

// adminAuthRateLimitEval 执行限流 Lua 脚本，返回当前计数。
// 设为包级变量以便测试时 mock 替换。
var adminAuthRateLimitEval = func(ctx context.Context, manager *redis.Manager, key string, window int) (int64, error) {
	result, err := manager.LuaWithContext(ctx, 1, rateLimitLuaScript, []string{key, strconv.Itoa(window)})
	if err != nil {
		return 0, err
	}
	count, _ := result.(int64)
	return count, nil
}

const (
	adminAuthRateLimitWindowDefault = 60
	adminAuthRateLimitMaxDefault    = 20
)

// AdminAuthRateLimit 限制认证端点的暴力破解尝试。
func (m middleware) AdminAuthRateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		cfg := config.Get().System.Admin.AuthRateLimit
		if !cfg.Enable {
			c.Next()
			return
		}

		window := cfg.WindowSeconds
		if window <= 0 {
			window = adminAuthRateLimitWindowDefault
		}

		maxRequests := cfg.MaxRequests
		if maxRequests <= 0 {
			maxRequests = adminAuthRateLimitMaxDefault
		}

		manager := m.redis["go-api"]
		if manager == nil {
			c.Next()
			return
		}

		// LuaWithContext 不自动添加 key 前缀，需手动拼接 manager.Prefix。
		key := fmt.Sprintf("%sadmin:auth:rate-limit:%s:%s", manager.Prefix, c.FullPath(), util.GetRealIP(c))
		count, err := adminAuthRateLimitEval(c.Request.Context(), manager, key, window)
		if err != nil {
			m.logger.Error(c.Request.Context(), "execute auth rate-limit lua script failed",
				zap.String("key", key), zap.Error(err))
			c.Next()
			return
		}

		if int(count) > maxRequests {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"code": http.StatusTooManyRequests,
				"msg":  "too many requests",
			})
			return
		}

		c.Next()
	}
}
