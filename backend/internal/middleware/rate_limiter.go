package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// TokenBucket 令牌桶结构
type TokenBucket struct {
	capacity     int64     // 桶容量
	tokens       int64     // 当前令牌数
	refillRate   int64     // 每秒补充令牌数
	lastRefill   time.Time // 上次补充时间
	mu           sync.Mutex
}

// NewTokenBucket 创建新的令牌桶
func NewTokenBucket(capacity, refillRate int64) *TokenBucket {
	return &TokenBucket{
		capacity:   capacity,
		tokens:     capacity,
		refillRate: refillRate,
		lastRefill: time.Now(),
	}
}

// TakeToken 尝试获取令牌
func (tb *TokenBucket) TakeToken() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(tb.lastRefill).Seconds()
	
	// 补充令牌
	tokensToAdd := int64(elapsed * float64(tb.refillRate))
	tb.tokens += tokensToAdd
	if tb.tokens > tb.capacity {
		tb.tokens = tb.capacity
	}
	tb.lastRefill = now

	// 尝试获取令牌
	if tb.tokens > 0 {
		tb.tokens--
		return true
	}
	return false
}

// RateLimiter 限流器结构
type RateLimiter struct {
	buckets map[string]*TokenBucket
	mu      sync.RWMutex
	capacity int64
	refillRate int64
}

// NewRateLimiter 创建新的限流器
func NewRateLimiter(capacity, refillRate int64) *RateLimiter {
	return &RateLimiter{
		buckets:    make(map[string]*TokenBucket),
		capacity:   capacity,
		refillRate: refillRate,
	}
}

// GetBucket 获取或创建令牌桶
func (rl *RateLimiter) GetBucket(key string) *TokenBucket {
	rl.mu.RLock()
	bucket, exists := rl.buckets[key]
	rl.mu.RUnlock()

	if exists {
		return bucket
	}

	rl.mu.Lock()
	defer rl.mu.Unlock()

	// 双重检查
	if bucket, exists := rl.buckets[key]; exists {
		return bucket
	}

	bucket = NewTokenBucket(rl.capacity, rl.refillRate)
	rl.buckets[key] = bucket
	return bucket
}

// CleanupExpiredBuckets 清理过期的令牌桶
func (rl *RateLimiter) CleanupExpiredBuckets() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	for key, bucket := range rl.buckets {
		bucket.mu.Lock()
		if now.Sub(bucket.lastRefill) > 10*time.Minute {
			delete(rl.buckets, key)
		}
		bucket.mu.Unlock()
	}
}

// 全局限流器实例
var (
	globalRateLimiter *RateLimiter
	userRateLimiter   *RateLimiter
	initOnce          sync.Once
)

// InitRateLimiters 初始化限流器
func InitRateLimiters() {
	initOnce.Do(func() {
		// 全局限流：每秒100个请求，桶容量200
		globalRateLimiter = NewRateLimiter(200, 100)
		
		// 用户限流：每秒10个请求，桶容量20
		userRateLimiter = NewRateLimiter(20, 10)

		// 启动清理协程
		go func() {
			ticker := time.NewTicker(5 * time.Minute)
			defer ticker.Stop()
			for range ticker.C {
				globalRateLimiter.CleanupExpiredBuckets()
				userRateLimiter.CleanupExpiredBuckets()
			}
		}()
	})
}

// GlobalRateLimit 全局限流中间件
func GlobalRateLimit() gin.HandlerFunc {
	InitRateLimiters()
	
	return func(c *gin.Context) {
		bucket := globalRateLimiter.GetBucket("global")
		
		if !bucket.TakeToken() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"error":   "请求过于频繁，请稍后再试",
				"code":    "RATE_LIMIT_EXCEEDED",
			})
			c.Abort()
			return
		}
		
		c.Next()
	}
}

// UserRateLimit 用户限流中间件
func UserRateLimit() gin.HandlerFunc {
	InitRateLimiters()
	
	return func(c *gin.Context) {
		// 获取用户标识（可以是用户ID、IP地址等）
		userID := getUserIdentifier(c)
		
		bucket := userRateLimiter.GetBucket(userID)
		
		if !bucket.TakeToken() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"error":   "用户请求过于频繁，请稍后再试",
				"code":    "USER_RATE_LIMIT_EXCEEDED",
			})
			c.Abort()
			return
		}
		
		c.Next()
	}
}

// ModelRateLimit 模型专用限流中间件
func ModelRateLimit(capacity, refillRate int64) gin.HandlerFunc {
	limiter := NewRateLimiter(capacity, refillRate)
	
	return func(c *gin.Context) {
		modelName := c.Param("name")
		if modelName == "" {
			modelName = "default"
		}
		
		bucket := limiter.GetBucket(modelName)
		
		if !bucket.TakeToken() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"error":   "模型请求过于频繁，请稍后再试",
				"code":    "MODEL_RATE_LIMIT_EXCEEDED",
				"model":   modelName,
			})
			c.Abort()
			return
		}
		
		c.Next()
	}
}

// getUserIdentifier 获取用户标识
func getUserIdentifier(c *gin.Context) string {
	// 优先使用用户ID
	if userID, exists := c.Get("user_id"); exists {
		if id, ok := userID.(string); ok {
			return "user:" + id
		}
	}
	
	// 回退到IP地址
	return "ip:" + c.ClientIP()
}

// GetRateLimitStatus 获取限流状态
func GetRateLimitStatus() gin.HandlerFunc {
	return func(c *gin.Context) {
		InitRateLimiters()
		
		userID := getUserIdentifier(c)
		userBucket := userRateLimiter.GetBucket(userID)
		globalBucket := globalRateLimiter.GetBucket("global")
		
		userBucket.mu.Lock()
		userTokens := userBucket.tokens
		userCapacity := userBucket.capacity
		userBucket.mu.Unlock()
		
		globalBucket.mu.Lock()
		globalTokens := globalBucket.tokens
		globalCapacity := globalBucket.capacity
		globalBucket.mu.Unlock()
		
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"user": gin.H{
					"available_tokens": userTokens,
					"capacity":         userCapacity,
					"usage_percent":    float64(userCapacity-userTokens) / float64(userCapacity) * 100,
				},
				"global": gin.H{
					"available_tokens": globalTokens,
					"capacity":         globalCapacity,
					"usage_percent":    float64(globalCapacity-globalTokens) / float64(globalCapacity) * 100,
				},
			},
		})
	}
}
