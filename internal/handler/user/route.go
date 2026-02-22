package user

import (
	"go-short/internal/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes 注册用户相关路由
func RegisterRoutes(r *gin.RouterGroup, handler *UserHandler) {
	userGroup := r.Group("/user")
	userGroup.Use(middleware.AuthMiddleware())
	{
		userGroup.GET("/profile", handler.GetProfile)
		userGroup.PUT("/profile", handler.UpdateProfile)
		userGroup.PUT("/password", handler.UpdatePassword)
	}
}
