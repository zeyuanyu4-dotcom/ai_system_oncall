package middleware

import (
	"strings"

	"ai_system_oncall/internal/response"
	"ai_system_oncall/pkg/jwt"

	"github.com/gin-gonic/gin"
)

// JWTAuth is a middleware for JWT authentication
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Fail(c, response.CodeUnauthorized, "未登录或Token无效")
			c.Abort()
			return
		}

		// Check Bearer prefix
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			response.Fail(c, response.CodeUnauthorized, "Token格式错误")
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Parse token
		claims, err := jwt.ParseToken(tokenString)
		if err != nil {
			response.Fail(c, response.CodeUnauthorized, "Token无效或已过期")
			c.Abort()
			return
		}

		// Store claims in context
		c.Set("claims", claims)
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)

		c.Next()
	}
}

// GetUserID gets user ID from context
func GetUserID(c *gin.Context) uint64 {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0
	}
	return userID.(uint64)
}

// GetUsername gets username from context
func GetUsername(c *gin.Context) string {
	username, exists := c.Get("username")
	if !exists {
		return ""
	}
	return username.(string)
}

// GetUserRole gets user role from context
func GetUserRole(c *gin.Context) string {
	role, exists := c.Get("role")
	if !exists {
		return ""
	}
	return role.(string)
}

// GetClaims gets JWT claims from context
func GetClaims(c *gin.Context) *jwt.Claims {
	claims, exists := c.Get("claims")
	if !exists {
		return nil
	}
	return claims.(*jwt.Claims)
}
