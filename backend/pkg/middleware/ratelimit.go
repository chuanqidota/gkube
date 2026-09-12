package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"gkube/pkg/response"
)

// 简易 IP 令牌桶限流器，用于登录接口防暴力破解。
// 后续可替换为 Redis 版本以支持多实例部署。

const (
	loginRateLimit  = 5              // 每窗口允许的最大请求数
	loginWindow     = 60 * time.Second // 滑动窗口大小
	cleanupInterval = 10 * time.Minute
)

// bucket 令牌桶状态
type bucket struct {
	mu       sync.Mutex
	tokens   int
	lastTime time.Time
}

var (
	loginBuckets sync.Map // map[string]*bucket
	cleanupOnce  sync.Once
)

// getBucket 获取或创建令牌桶，返回时持有锁（调用方负责解锁）。
func getBucket(key string) *bucket {
	val, _ := loginBuckets.LoadOrStore(key, &bucket{
		tokens:   loginRateLimit,
		lastTime: time.Now(),
	})
	b := val.(*bucket)
	b.mu.Lock()

	now := time.Now()
	elapsed := now.Sub(b.lastTime)
	// 按时间恢复令牌
	replenish := int(elapsed / (loginWindow / time.Duration(loginRateLimit)))
	if replenish > 0 {
		b.tokens += replenish
		if b.tokens > loginRateLimit {
			b.tokens = loginRateLimit
		}
		b.lastTime = now
	}
	return b // 调用方必须 b.mu.Unlock()
}

// startCleanup 启动后台清理协程（只执行一次）。
func startCleanup() {
	cleanupOnce.Do(func() {
		go func() {
			ticker := time.NewTicker(cleanupInterval)
			defer ticker.Stop()
			for range ticker.C {
				loginBuckets.Range(func(key, val any) bool {
					b := val.(*bucket)
					b.mu.Lock()
					idle := time.Since(b.lastTime) > loginWindow*2
					b.mu.Unlock()
					if idle {
						loginBuckets.Delete(key)
					}
					return true
				})
			}
		}()
	})
}

// RateLimitLogin 对登录接口进行 IP 级别限流的 Gin 中间件。
func RateLimitLogin() gin.HandlerFunc {
	startCleanup()

	return func(c *gin.Context) {
		ip := c.ClientIP()
		b := getBucket(ip)

		if b.tokens <= 0 {
			b.mu.Unlock()
			response.FailWithStatus(c, http.StatusTooManyRequests, "请求过于频繁，请稍后再试")
			return
		}
		b.tokens--
		b.mu.Unlock()

		c.Next()
	}
}
