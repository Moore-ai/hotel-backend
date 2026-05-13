package middleware

import (
	"context"
	"strings"

	"github.com/gin-gonic/gin"
	"hotel-backend/internal/database"
	"hotel-backend/pkg/errcode"
	"hotel-backend/pkg/jwt"
)

var JWTSecret string

func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			c.Abort()
			errcode.Write(c, errcode.ErrUnauthorized)
			return
		}
		tokenStr := strings.TrimPrefix(header, "Bearer ")

		ctx := context.Background()
		exists, _ := database.RDB.Exists(ctx, "blacklist:"+tokenStr).Result()
		if exists > 0 {
			c.Abort()
			errcode.Write(c, errcode.ErrTokenInvalid)
			return
		}

		claims, err := jwt.ParseToken(tokenStr, JWTSecret)
		if err != nil {
			c.Abort()
			errcode.Write(c, errcode.ErrTokenExpired)
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("role", claims.Role)
		c.Next()
	}
}
