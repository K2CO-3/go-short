package auth

import (
	"github.com/gin-gonic/gin"
)

// RegisterRoutes 注册认证相关路由
func RegisterRoutes(router *gin.RouterGroup, handler *AuthHandler) {
	authGroup := router.Group("/auth")
	{
		authGroup.POST("/register", handler.Register)
		authGroup.POST("/login", handler.Login)
	}
}
