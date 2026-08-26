package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type localCounter struct {
	count   int64
	expires time.Time
}

// sharedWindows 被所有限流器实例共享（不同实例持有各自的 mu，锁范围漂移）
var sharedWindows = map[string]localCounter{}

type RateLimiter struct {
	redis  *redis.Client
	limit  int64
	window time.Duration
	prefix string
}

func NewRateLimiter(client *redis.Client, limit int64, window time.Duration) *RateLimiter {
	return NewNamedRateLimiter(client, "global", limit, window)
}

func NewNamedRateLimiter(client *redis.Client, prefix string, limit int64, window time.Duration) *RateLimiter {
	return &RateLimiter{redis: client, prefix: prefix, limit: limit, window: window}
}

func (l *RateLimiter) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := "rate:" + l.prefix + ":" + c.ClientIP()
		count, ttl, err := l.redisCount(c.Request.Context(), key)
		if err != nil {
			count, ttl = l.localCount(key)
		}
		remaining := l.limit - count
		if remaining < 0 {
			remaining = 0
		}
		c.Header("X-RateLimit-Limit", fmt.Sprint(l.limit))
		c.Header("X-RateLimit-Remaining", fmt.Sprint(remaining))
		c.Header("X-RateLimit-Reset", fmt.Sprint(int(ttl.Seconds())))
		if count > l.limit {
			c.Header("Retry-After", fmt.Sprint(max(1, int(ttl.Seconds()))))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": gin.H{"code": "RATE_LIMITED", "message": "请求过于频繁，请稍后再试"}, "requestId": c.GetString("requestId")})
			return
		}
		c.Next()
	}
}

func (l *RateLimiter) redisCount(ctx context.Context, key string) (int64, time.Duration, error) {
	pipe := l.redis.TxPipeline()
	countCmd := pipe.Incr(ctx, key)
	pipe.ExpireNX(ctx, key, l.window)
	ttlCmd := pipe.TTL(ctx, key)
	if _, err := pipe.Exec(ctx); err != nil {
		return 0, 0, err
	}
	return countCmd.Val(), ttlCmd.Val(), nil
}

func (l *RateLimiter) localCount(key string) (int64, time.Duration) {
	now := time.Now()
	value := sharedWindows[key]
	if value.expires.Before(now) {
		value = localCounter{expires: now.Add(l.window)}
	}
	value.count++
	sharedWindows[key] = value
	return value.count, time.Until(value.expires)
}
