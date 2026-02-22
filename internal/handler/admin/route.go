package admin

import (
	"go-short/internal/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes 注册管理员相关路由
func RegisterRoutes(router *gin.RouterGroup, handler *AdminHandler) {
	adminGroup := router.Group("/admin")
	adminGroup.Use(middleware.AdminMiddleware())
	{
		adminGroup.POST("/createUser", handler.CreateUser)
		adminGroup.GET("/getUserList", handler.GetUsers)
		adminGroup.DELETE("/delete/:userID", handler.DeleteUser)
		adminGroup.PUT("/unactivateUser/:userID", handler.UnactiveUser)
		adminGroup.PUT("/activateUser/:userID", handler.ActiveUser)
		adminGroup.PUT("/activateLink/:linkID", handler.ActiveLink)
		adminGroup.PUT("/unactivateLink/:linkID", handler.UnactiveLink)
	}
}
