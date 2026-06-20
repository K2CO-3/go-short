package main

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof" // 零埋点 pprof，import 即生效
	"os"
	"strings"
	"time"

	"go-short/internal/bloom"
	"go-short/internal/metrics"
	"go-short/internal/mq"
	"go-short/internal/repository/impl/local"
	"go-short/internal/repository/impl/postgresql"
	"go-short/internal/repository/impl/redis"
	"go-short/internal/service"

	"github.com/gin-gonic/gin"
	redisclient "github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
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

	kafkaWriter := mq.NewAccessLogWriter()
	defer kafkaWriter.Close()

	// 2. 初始化 Repository
	linkRepo := postgresql.NewLinkRepository(db)
	userRepo := postgresql.NewUserRepository(db)
	accessLogRepo := postgresql.NewAccessLogRepository(db)
	redisRepo := redis.NewRedisRepository(rdb, nil)

	// 初始化本地缓存（TTL 5min + LRU 最多 10000 条）
	localCache := local.NewLocalCache(5*time.Minute, 10*time.Minute, 10000)
	log.Println("✅ Local cache initialized")

	// 布隆过滤器防 Redis 缓存穿透（预期 100w 短码，1% 误判率）
	shortCodeBloom := bloom.NewShortCodeBloom(1_000_000, 0.01)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()
		n, err := shortCodeBloom.LoadFromDB(ctx, db)
		if err != nil {
			log.Printf("⚠️ Bloom filter load failed: %v (penetration protection disabled)", err)
		} else {
			log.Printf("✅ Bloom filter loaded %d short codes", n)
		}
	}()

	// Kafka 订阅：缓存失效 + 布隆增量（同一 topic，按 event 区分；每实例独立 group 实现广播）
	go func() {
		groupID := mq.CacheInvalidateConsumerGroupID()
		reader := mq.NewCacheInvalidateReader(groupID)
		defer reader.Close()
		ctx := context.Background()
		log.Printf("📡 Redirect sync (cache + bloom) Kafka consumer group: %s", groupID)
		for {
			msg, err := reader.FetchMessage(ctx)
			if err != nil {
				log.Println("redirect sync Kafka read error, retrying in 2s...", err)
				time.Sleep(2 * time.Second)
				continue
			}
			var payload struct {
				Code  string `json:"code"`
				Event string `json:"event"`
			}
			if err := json.Unmarshal(msg.Value, &payload); err != nil || payload.Code == "" {
				log.Printf("redirect sync skip invalid payload: %s err=%v", string(msg.Value), err)
				_ = reader.CommitMessages(ctx, msg)
				continue
			}
			switch payload.Event {
			case mq.EventBloomAdd:
				shortCodeBloom.Add(payload.Code)
			case mq.EventCacheInvalidate, "":
				localCache.Delete("short:" + payload.Code)
			default:
				log.Printf("redirect sync unknown event %q for code=%s, evicting local cache only", payload.Event, payload.Code)
				localCache.Delete("short:" + payload.Code)
			}
			if err := reader.CommitMessages(ctx, msg); err != nil {
				log.Println("redirect sync Kafka commit error:", err)
			}
		}
	}()

	// 3. 初始化 Service（redirect 只读，无需 cacheInvalidator）
	linkService := service.NewLinkService(db, linkRepo, userRepo, accessLogRepo, nil, nil)

	// 4. 启动 pprof（零埋点，import 即注册，6060 端口）
	go func() { _ = http.ListenAndServe(":6060", nil) }()

	// 5. 初始化 Gin 引擎
	r := gin.Default()
	// 仅信任反向代理（默认本机 Nginx），避免任意来源伪造 X-Forwarded-For
	if err := r.SetTrustedProxies(resolveTrustedProxies()); err != nil {
		log.Fatal("Failed to set trusted proxies:", err)
	}

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

		// Step 2: 布隆过滤器防穿透（一定不存在则直返 404，不打 Redis/DB）
		if !found && shortCodeBloom.DefinitelyNotExist(code) {
			c.String(404, "Link not found or expired")
			return
		}

		// Step 3: 查 Redis
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

		// Step 4: Redis 未命中，查 PostgreSQL（回源，分布式锁防缓存击穿）
		if !found {
			lockKey := redis.LockKeyForCode(code)
			token, gotLock := redisRepo.TryLock(ctx, lockKey, redis.LockTTLSeconds*time.Second)
			if gotLock {
				defer func() { _ = redisRepo.Unlock(context.Background(), lockKey, token) }()
				if url, err := redisRepo.GetLinkFromCache(ctx, code); err == nil {
					longURL = url
					found = true
					localCache.Set(cacheKey, longURL)
				}
			}
			if !found && !gotLock {
				for i := 0; i < 20; i++ {
					time.Sleep(50 * time.Millisecond)
					if url, err := redisRepo.GetLinkFromCache(ctx, code); err == nil {
						longURL = url
						found = true
						localCache.Set(cacheKey, longURL)
						break
					}
				}
			}
			if !found {
				startTime := time.Now()
				link, dbErr := linkService.GetLinkByCodeForRedirect(ctx, code)
				postgresDuration := time.Since(startTime)
				if dbErr != nil {
					metrics.RecordPostgres(false, postgresDuration)
					c.String(404, "Link not found or expired")
					return
				}
				longURL = link.OriginalURL
				metrics.RecordPostgres(true, postgresDuration)
				shortCodeBloom.Add(code) // DB 命中则加入布隆（新建链接首次访问）
				localCache.Set(cacheKey, longURL)
				redisRepo.CacheLink(ctx, code, longURL, time.Hour)
			}
		}

		// Step 5: 异步发送访问日志到 Kafka
		clientIP := realClientIP(c)
		go func(code, ip, ua string) {
			bgCtx := context.Background()
			logData := map[string]any{
				"code": code,
				"ip":   ip,
				"ua":   ua,
				"ts":   time.Now().Unix(),
			}
			dataBytes, _ := json.Marshal(logData)
			_ = kafkaWriter.WriteMessages(bgCtx, kafka.Message{Value: dataBytes})
			rdb.Incr(bgCtx, "stats:visits:"+code)
		}(code, clientIP, c.Request.UserAgent())

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

func resolveTrustedProxies() []string {
	// 可通过 TRUSTED_PROXIES 覆盖，例如：
	// TRUSTED_PROXIES=127.0.0.1,::1,10.0.0.0/8
	raw := strings.TrimSpace(os.Getenv("TRUSTED_PROXIES"))
	if raw == "" {
		return []string{"127.0.0.1", "::1"}
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		return []string{"127.0.0.1", "::1"}
	}
	return out
}

func realClientIP(c *gin.Context) string {
	ip := strings.TrimSpace(c.ClientIP())
	if ip != "" {
		return ip
	}
	host, _, err := net.SplitHostPort(strings.TrimSpace(c.Request.RemoteAddr))
	if err == nil && host != "" {
		return host
	}
	return strings.TrimSpace(c.Request.RemoteAddr)
}
