package system

import (
	"context"
	"github.com/sk-pkg/redis"
)

const userCachePrefix = "admin:system:user"

func clearUserCache(ctx context.Context, redis *redis.Manager) error {
	return redis.BatchDelWithContext(ctx, userCachePrefix)
}
