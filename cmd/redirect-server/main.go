package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	_ "net/http/pprof" // 零埋点 pprof，import 即生效
	"time"

	"go-short/internal/metrics"
	"go-short/internal/repository/impl/local"
	"go-short/internal/repository/impl/postgresql"
	"go-short/internal/repository/impl/redis"
	"go-short/internal/service"

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
	userRepo := postgresql.NewUserRepository(db)
	accessLogRepo := postgresql.NewAccessLogRepository(db)
	redisRepo := redis.NewRedisRepository(rdb)

	// 初始化本地缓存
	// 默认TTL: 5分钟（本地缓存时间较短，避免数据不一致）
	// 清理间隔: 10分钟（定期清理过期项）
	localCache := local.NewLocalCache(5*time.Minute, 10*time.Minute)
	log.Println("✅ Local cache initialized")

	// 3. 初始化 Service
	linkService := service.NewLinkService(db, linkRepo, userRepo, accessLogRepo)

	// 4. 启动 pprof（零埋点，import 即注册，6060 端口）
	go func() { _ = http.ListenAndServe(":6060", nil) }()

	// 5. 初始化 Gin 引擎
	r := gin.Default()

	// 6. 核心跳转路由（带延迟统计）
	r.GET("/code/:code", func(c *gin.Context) {
		code := c.Param("code")
		if code == "" {
			c.String(400, "Bad Request")
			return
		}

		ctx := c.Request.Context()
		var longURL string
		var found bool

		// 三级缓存查询：本地缓存 -> Redis -> PostgreSQL

		// Step 1: 查本地缓存（最快）
		cacheKey := "short:" + code
		startTime := time.Now()
		if url, ok := localCache.Get(cacheKey); ok {
			longURL = url
			found = true
			// 记录本地缓存命中延迟
			metrics.RecordLocalCache(true, time.Since(startTime))
		} else {
			// 记录本地缓存未命中延迟
			metrics.RecordLocalCache(false, time.Since(startTime))
		}

		// Step 2: 本地缓存未命中，查 Redis
		if !found {
			startTime = time.Now()
			cachedURL, err := redisRepo.GetLinkFromCache(ctx, code)
			redisDuration := time.Since(startTime)

			if err == nil {
				// Redis 命中
				longURL = cachedURL
				found = true
				metrics.RecordRedis(true, redisDuration)
				// 回填本地缓存（使用较短的TTL，5分钟）
				localCache.Set(cacheKey, longURL)
			} else if err == redisclient.Nil {
				// Redis 未命中（正常情况）
				metrics.RecordRedis(false, redisDuration)
			} else {
				// Redis 错误（非未命中），记录日志但继续降级到数据库
				metrics.RecordRedis(false, redisDuration)
				log.Printf("Redis error for code %s: %v, falling back to database", code, err)
			}
		}

		// Step 3: Redis 未命中，查 PostgreSQL（回源）
		if !found {
			startTime = time.Now()
			link, dbErr := linkService.GetLinkByCodeForRedirect(ctx, code)
			postgresDuration := time.Since(startTime)

			if dbErr != nil {
				metrics.RecordPostgres(false, postgresDuration)
				c.String(404, "Link not found or expired")
				return
			}

			longURL = link.OriginalURL
			metrics.RecordPostgres(true, postgresDuration)

			// Step 4: 回填缓存（本地缓存 + Redis）
			// 本地缓存：5分钟TTL
			localCache.Set(cacheKey, longURL)
			// Redis：1小时TTL（较长的过期时间）
			redisRepo.CacheLink(ctx, code, longURL, time.Hour)
		}

		// Step 5: 【异步】发送访问日志到 Redis 队列
		// 使用 go 协程 + context.Background()，避免 302 返回后 request context 被取消导致 RPush 失败
		go func(code, ip, ua string) {
			bgCtx := context.Background()
			logData := map[string]any{
				"code": code,
				"ip":   ip,
				"ua":   ua,
				"ts":   time.Now().Unix(),
			}
			dataBytes, _ := json.Marshal(logData)

			rdb.RPush(bgCtx, "access_logs", dataBytes)
			rdb.Incr(bgCtx, "stats:visits:"+code)
		}(code, c.ClientIP(), c.Request.UserAgent())

		// Step 6: 302 重定向
		c.Redirect(http.StatusFound, longURL)
	})

	// 7. 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// 8. Metrics 端点：展示延迟统计
	r.GET("/metrics", func(c *gin.Context) {
		stats := metrics.FormatStats()
		c.JSON(200, gin.H{
			"success": true,
			"data":    stats,
			"note":    "延迟单位：毫秒(ms)，hit_rate 单位：百分比(%)",
		})
	})

	log.Println("🚀 Redirect Server running on :8080")
	log.Println("📊 Metrics endpoint: http://localhost:8080/metrics")
	r.Run(":8080")
}
