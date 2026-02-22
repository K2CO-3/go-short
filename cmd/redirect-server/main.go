package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"go-short/internal/repository/impl/postgresql"
	"go-short/internal/repository/impl/redis"

	"github.com/gin-gonic/gin"
	redisclient "github.com/redis/go-redis/v9"
)

func main() {
	// 生产模式，减少日志输出提升性能
	gin.SetMode(gin.ReleaseMode)

	// 1. 初始化资源
	db, err := postgresql.NewPostgresClient()
	if err != nil {
		log.Fatal("Failed to connect to DB:", err)
	}

	rdb, err := redis.NewRedisClient()
	if err != nil {
		log.Fatal("Failed to connect to Redis:", err)
	}

	// 2. 初始化 Repository
	linkRepo := postgresql.NewLinkRepository(db)
	redisRepo := redis.NewRedisRepository(rdb)

	// 3. 初始化 Gin 引擎
	r := gin.Default()

	// 4. 核心跳转路由
	r.GET("/code/:code", func(c *gin.Context) {
		code := c.Param("code")
		if code == "" {
			c.String(400, "Bad Request")
			return
		}

		ctx := c.Request.Context()
		var longURL string

		// Step 1: 查 Redis 缓存
		cachedURL, err := redisRepo.GetLinkFromCache(ctx, code)
		if err == redisclient.Nil {
			// 缓存未命中
			longURL = cachedURL
		} else if err != nil {
			// Redis 错误
			c.String(500, "Redis error")
			return
		} else {
			// Step 2: 缓存未命中，查数据库 (回源)
			link, dbErr := linkRepo.GetLinkByCode(ctx, nil, code)
			if dbErr != nil {
				c.String(404, "Link not found or expired")
				return
			}

			// 检查链接是否过期
			if link.ExpiresAt != nil && link.ExpiresAt.Before(time.Now()) {
				c.String(404, "Link expired")
				return
			}

			longURL = link.OriginalURL

			// Step 3: 回写 Redis 缓存 (设置1小时过期)
			redisRepo.CacheLink(ctx, code, longURL, time.Hour)
		}

		// Step 4: 【异步】发送访问日志到 Redis 队列
		// 使用 go 协程，不阻塞 HTTP 跳转
		go func(code, ip, ua string) {
			logData := map[string]interface{}{
				"code": code,
				"ip":   ip,
				"ua":   ua,
				"ts":   time.Now().Unix(),
			}
			dataBytes, _ := json.Marshal(logData)

			// 推送到名为 "access_logs" 的 List
			rdb.RPush(ctx, "access_logs", dataBytes)

			// 简单的实时计数器 (用于 Dashboard 快速展示)
			rdb.Incr(ctx, "stats:visits:"+code)
		}(code, c.ClientIP(), c.Request.UserAgent())

		// Step 5: 302 重定向
		c.Redirect(http.StatusFound, longURL)
	})

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	log.Println("🚀 Redirect Server running on :8080")
	r.Run(":8080")
}
