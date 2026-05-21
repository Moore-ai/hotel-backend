package middleware

import (
	"fmt"
	"log"
	"time"

	"hotel-backend/config"
	"hotel-backend/internal/database"
	"hotel-backend/pkg/errcode"

	"github.com/gin-gonic/gin"
)

// chatRateLimitCfg AI 对话速率限制配置，在 main.go 中设置
var chatRateLimitCfg *config.RateLimitConfig

// InitChatRateLimit 在 main.go 中调用，设置速率限制配置
func InitChatRateLimit(cfg *config.RateLimitConfig) {
	chatRateLimitCfg = cfg
}

// ChatRateLimit 对 AI 智能管家对话接口进行速率限制
func ChatRateLimit() gin.HandlerFunc {
	cfg := chatRateLimitCfg
	if !cfg.Enabled {
		return func(c *gin.Context) { c.Next() }
	}

	rpm := cfg.MaxRequestsPerMinute
	if rpm <= 0 {
		rpm = 10
	}

	return func(c *gin.Context) {
		userID := c.GetUint("user_id")
		key := fmt.Sprintf("ratelimit:chat:%d", userID)
		ctx := c.Request.Context()

		// 使用事务保证 INCR 和 EXPIRE 原子性
		pipe := database.RDB.TxPipeline()
		incr := pipe.Incr(ctx, key)
		pipe.Expire(ctx, key, time.Minute)
		_, err := pipe.Exec(ctx)
		if err != nil {
			log.Printf("Rate limit check failed: %v", err)
			c.Next()
			return
		}

		if incr.Val() > int64(rpm) {
			errcode.Write(c, errcode.ErrRateLimited)
			c.Abort()
			return
		}

		c.Next()
	}
}
