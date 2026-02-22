package link

import (
	"go-short/internal/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes 注册链接相关路由
func RegisterRoutes(r *gin.RouterGroup, handler *LinkHandler) {
	linksGroup := r.Group("/links")
	linksGroup.Use(middleware.AuthMiddleware())
	{
		linksGroup.POST("", handler.Create)
		linksGroup.GET("", handler.GetLinks)
		linksGroup.GET("/GetLinksByAlias", handler.GetLinksByAlias)
		linksGroup.DELETE("/:id", handler.Delete)
	}
}
