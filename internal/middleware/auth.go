// internal/middleware/auth.go
package middleware

import (
	"go-short/internal/util"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware 通用认证中间件：解析 token 并设置用户信息（不检查角色）
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := c.GetHeader("Authorization")

		// 检查 Authorization header 格式
		if tokenStr == "" {
			c.AbortWithStatusJSON(401, gin.H{"error": "Authorization header required"})
			return
		}

		// 提取Token (Bearer <token>)
		if len(tokenStr) <= 7 || !strings.HasPrefix(tokenStr, "Bearer ") {
			c.AbortWithStatusJSON(401, gin.H{"error": "Invalid authorization header format"})
			return
		}

		token := tokenStr[7:]
		claims, err := util.ParseToken(token)
		if err != nil {
			// token 解析失败（可能是过期、签名错误等）
			c.AbortWithStatusJSON(401, gin.H{"error": "Invalid or expired token"})
			return
		}

		if claims == nil {
			c.AbortWithStatusJSON(401, gin.H{"error": "Invalid token claims"})
			return
		}

		// 设置用户信息到上下文
		c.Set("uid", claims.UserID.String())
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Next()
	}
}
