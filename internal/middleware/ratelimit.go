package middleware

import (
	_ "embed"
	"fmt"
	"log"
	"strconv"
	"time"

	"hotel-backend/config"
	"hotel-backend/internal/database"
	"hotel-backend/pkg/errcode"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

//go:embed token_bucket.lua
var tokenBucketScriptSrc string

const (
	rateLimitFixedWindow   = "fixed_window"
	rateLimitTokenBucket   = "token_bucket"
	rateLimitSlidingWindow = "sliding_window"
)

var (
	chatRateLimitCfg     *config.RateLimitConfig
	tokenBucketScript    = redis.NewScript(tokenBucketScriptSrc)
)

func InitChatRateLimit(cfg *config.RateLimitConfig) {
	chatRateLimitCfg = cfg
}

func ChatRateLimit() gin.HandlerFunc {
	cfg := chatRateLimitCfg
	if cfg == nil || !cfg.Enabled {
		return func(c *gin.Context) { c.Next() }
	}

	rpm := cfg.MaxRequestsPerMinute
	if rpm <= 0 {
		rpm = 10
	}

	var limiter func(ctx *gin.Context, userID uint) (ok bool, retryAfter int)
	switch cfg.Algorithm {
	case rateLimitTokenBucket:
		limiter = tokenBucket(rpm)
	case rateLimitSlidingWindow:
		limiter = slidingWindow(rpm)
	case rateLimitFixedWindow, "":
		limiter = fixedWindow(rpm)
	default:
		log.Printf("Rate limit: unknown algorithm %q, falling back to fixed_window", cfg.Algorithm)
		limiter = fixedWindow(rpm)
	}

	return func(c *gin.Context) {
		userID := c.GetUint("user_id")
		if cfg.IsWhitelisted(userID) {
			c.Next()
			return
		}
		ok, retryAfter := limiter(c, userID)
		if !ok {
			c.Header("Retry-After", strconv.Itoa(retryAfter))
			errcode.Write(c, errcode.ErrRateLimited)
			c.Abort()
			return
		}
		c.Next()
	}
}

func fixedWindow(rpm int) func(*gin.Context, uint) (bool, int) {
	return func(c *gin.Context, userID uint) (bool, int) {
		key := fmt.Sprintf("ratelimit:fw:%d", userID)
		ctx := c.Request.Context()

		pipe := database.RDB.TxPipeline()
		incr := pipe.Incr(ctx, key)
		pipe.Expire(ctx, key, time.Minute)
		_, err := pipe.Exec(ctx)
		if err != nil {
			log.Printf("Rate limit (fixed window) failed: %v", err)
			return true, 0
		}

		if incr.Val() > int64(rpm) {
			ttl, _ := database.RDB.TTL(ctx, key).Result()
			return false, int(ttl.Seconds())
		}
		return true, 0
	}
}

func tokenBucket(rpm int) func(*gin.Context, uint) (bool, int) {
	return func(c *gin.Context, userID uint) (bool, int) {
		key := fmt.Sprintf("ratelimit:tb:%d", userID)
		tsKey := key + ":ts"
		ctx := c.Request.Context()

		ret, err := tokenBucketScript.Run(ctx, database.RDB, []string{key, tsKey},
			time.Now().UnixMilli(), rpm, 60000).Int()
		if err != nil {
			log.Printf("Rate limit (token bucket) failed: %v", err)
			return true, 0
		}

		if ret != 1 {
			ttl, _ := database.RDB.TTL(ctx, key).Result()
			return false, max(int(ttl.Seconds()), 1)
		}
		return true, 0
	}
}

func slidingWindow(rpm int) func(*gin.Context, uint) (bool, int) {
	return func(c *gin.Context, userID uint) (bool, int) {
		key := fmt.Sprintf("ratelimit:sw:%d", userID)
		ctx := c.Request.Context()
		now := time.Now().UnixMilli()
		cutoff := now - time.Minute.Milliseconds()

		pipe := database.RDB.TxPipeline()
		pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", cutoff))
		pipe.ZAdd(ctx, key, redis.Z{Score: float64(now), Member: fmt.Sprintf("%d", now)})
		zcard := pipe.ZCard(ctx, key)
		pipe.Expire(ctx, key, time.Minute)
		_, err := pipe.Exec(ctx)
		if err != nil {
			log.Printf("Rate limit (sliding window) failed: %v", err)
			return true, 0
		}

		if zcard.Val() > int64(rpm) {
			ttl, _ := database.RDB.TTL(ctx, key).Result()
			return false, int(ttl.Seconds())
		}
		return true, 0
	}
}
