package middleware

import (
	"ai_system_oncall/internal/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Recovery is a middleware for recovering from panics
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				zap.L().Error("Panic recovered",
					zap.Any("error", err),
					zap.String("path", c.Request.URL.Path),
				)

				response.Fail(c, response.CodeInternalError, "服务器内部错误")
				c.Abort()
			}
		}()
		c.Next()
	}
}
