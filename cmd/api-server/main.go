package main

import (
	"log"

	"go-short/internal/handler/admin"
	"go-short/internal/handler/auth"
	"go-short/internal/handler/link"
	"go-short/internal/handler/user"
	"go-short/internal/handler/validator"
	"go-short/internal/middleware"
	"go-short/internal/repository/impl/postgresql"
	"go-short/internal/repository/impl/redis"
	"go-short/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func main() {
	// 1. 初始化数据库连接
	db, err := postgresql.NewPostgresClient()
	if err != nil {
		log.Fatal("Failed to connect to DB:", err)
	}

	rdb, err := redis.NewRedisClient()
	if err != nil {
		log.Fatal("Failed to connect to Redis:", err)
	}
	_ = rdb // Redis 客户端保留，供后续使用

	// 2. 初始化 Repository
	userRepo := postgresql.NewUserRepository(db)
	linkRepo := postgresql.NewLinkRepository(db)
	accessLogRepo := postgresql.NewAccessLogRepository(db)

	// 3. 初始化 Service
	userService := service.NewUserService(db, userRepo)
	linkService := service.NewLinkService(db, linkRepo, userRepo, accessLogRepo)
	adminService := service.NewAdminService(db, linkRepo, userRepo, accessLogRepo)

	// 4. 初始化 Handler
	authHandler := auth.NewAuthHandler(userService)
	linkHandler := link.NewLinkHandler(linkService)
	adminHandler := admin.NewAdminHandler(adminService)
	userHandler := user.NewUserHandler(userService)

	// 5. 将自定义验证规则注册到Gin的默认validator
	// 必须在创建 Gin 引擎之前调用，这样 ShouldBindJSON 等绑定方法才能使用自定义验证规则
	validator.SetupGinValidator()

	// 6. 初始化 Gin 引擎
	r := gin.Default()

	// 简单的 CORS 中间件
	r.Use(middleware.CORS())

	// 7. 注册路由
	api := r.Group("/api/v1")
	auth.RegisterRoutes(api, authHandler)
	link.RegisterRoutes(api, linkHandler)
	user.RegisterRoutes(api, userHandler)
	admin.RegisterRoutes(api, adminHandler)

	// 8. 测试接口（仅鉴权 + 数据库查询）
	testGroup := api.Group("/test")
	testGroup.Use(middleware.AuthMiddleware())
	{
		testGroup.GET("/ping", func(c *gin.Context) {
			// 从中间件获取用户ID
			uidStr, exists := c.Get("uid")
			if !exists {
				c.JSON(401, gin.H{"error": "用户信息未找到"})
				return
			}

			userID, err := uuid.Parse(uidStr.(string))
			if err != nil {
				c.JSON(400, gin.H{"error": "无效的用户ID"})
				return
			}

			// 查询数据库获取用户完整信息
			user, err := userService.GetUserByUserID(c, userID)
			if err != nil {
				c.JSON(500, gin.H{
					"success": false,
					"error":   "查询用户信息失败",
					"message": err.Error(),
				})
				return
			}

			// 返回用户信息
			email := ""
			if user.Email != nil {
				email = *user.Email
			}

			c.JSON(200, gin.H{
				"success":    true,
				"message":    "认证成功，数据库查询成功",
				"user_id":    user.ID.String(),
				"username":   user.Username,
				"email":      email,
				"role":       user.Role,
				"status":     user.Status,
				"link_count": user.LinkCount,
				"created_at": user.CreatedAt.Format("2006-01-02T15:04:05Z"),
				"updated_at": user.UpdatedAt.Format("2006-01-02T15:04:05Z"),
			})
		})
	}

	log.Println("🚀 API Server running on :8080")
	r.Run(":8080")
}
