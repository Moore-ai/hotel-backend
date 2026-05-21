package middleware

import (
	_ "embed"
	"fmt"
	"log"
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

var chatRateLimitCfg *config.RateLimitConfig

func InitChatRateLimit(cfg *config.RateLimitConfig) {
	chatRateLimitCfg = cfg
}

var tokenBucketScript = redis.NewScript(tokenBucketScriptSrc)

func ChatRateLimit() gin.HandlerFunc {
	cfg := chatRateLimitCfg
	if cfg == nil || !cfg.Enabled {
		return func(c *gin.Context) { c.Next() }
	}

	rpm := cfg.MaxRequestsPerMinute
	if rpm <= 0 {
		rpm = 10
	}

	var limiter func(ctx *gin.Context, userID uint) bool
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
		if !limiter(c, userID) {
			errcode.Write(c, errcode.ErrRateLimited)
			c.Abort()
			return
		}
		c.Next()
	}
}

func fixedWindow(rpm int) func(*gin.Context, uint) bool {
	return func(c *gin.Context, userID uint) bool {
		key := fmt.Sprintf("ratelimit:fw:%d", userID)
		ctx := c.Request.Context()

		pipe := database.RDB.TxPipeline()
		incr := pipe.Incr(ctx, key)
		pipe.Expire(ctx, key, time.Minute)
		_, err := pipe.Exec(ctx)
		if err != nil {
			log.Printf("Rate limit (fixed window) failed: %v", err)
			return true
		}
		return incr.Val() <= int64(rpm)
	}
}

func tokenBucket(rpm int) func(*gin.Context, uint) bool {
	return func(c *gin.Context, userID uint) bool {
		key := fmt.Sprintf("ratelimit:tb:%d", userID)
		tsKey := key + ":ts"
		ctx := c.Request.Context()

		ret, err := tokenBucketScript.Run(ctx, database.RDB, []string{key, tsKey},
			time.Now().UnixMilli(), rpm, 60000).Int()
		if err != nil {
			log.Printf("Rate limit (token bucket) failed: %v", err)
			return true
		}
		return ret == 1
	}
}

func slidingWindow(rpm int) func(*gin.Context, uint) bool {
	return func(c *gin.Context, userID uint) bool {
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
			return true
		}

		return zcard.Val() <= int64(rpm)
	}
}
