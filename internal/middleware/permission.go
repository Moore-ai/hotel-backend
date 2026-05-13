package middleware

import (
	"github.com/gin-gonic/gin"
	"hotel-backend/pkg/errcode"
)

func RequireRoles(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			c.Abort()
			errcode.Write(c, errcode.ErrForbidden)
			return
		}
		for _, r := range roles {
			if role.(string) == r {
				c.Next()
				return
			}
		}
		c.Abort()
		errcode.Write(c, errcode.ErrForbidden)
	}
}
