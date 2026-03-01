package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strconv"
	"time"

	"go-short/internal/model"
	"go-short/internal/repository"
	"go-short/internal/repository/impl/postgresql"
	"go-short/internal/repository/impl/redis"
)

// LogPayload 对应 Redirect Server 发送的 JSON 结构
type LogPayload struct {
	Code string `json:"code"`
	IP   string `json:"ip"`
	UA   string `json:"ua"`
	TS   int64  `json:"ts"`
}

func main() {
	db, err := postgresql.NewPostgresClient()
	if err != nil {
		log.Fatal("Failed to connect to DB:", err)
	}

	rdb, err := redis.NewRedisClient()
	if err != nil {
		log.Fatal("Failed to connect to Redis:", err)
	}

	linkRepo := postgresql.NewLinkRepository(db)
	accessLogRepo := postgresql.NewAccessLogRepository(db)

	poolSize := 5
	if s := os.Getenv("WORKER_POOL_SIZE"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 {
			poolSize = n
		}
	}

	log.Printf("👷 Worker started, pool size=%d, waiting for logs...\n", poolSize)

	ctx := context.Background()
	jobs := make(chan LogPayload, poolSize*2)

	// 协程池：多个 worker 并发消费
	for i := 0; i < poolSize; i++ {
		go func(id int) {
			for payload := range jobs {
				processLog(ctx, id, payload, linkRepo, accessLogRepo)
			}
		}(i)
	}

	// 主 goroutine：BLPop 入队
	for {
		result, err := rdb.BLPop(ctx, 0, "access_logs").Result()
		if err != nil {
			log.Println("Redis connect error, retrying in 5s...", err)
			time.Sleep(5 * time.Second)
			continue
		}

		rawJSON := result[1]
		var payload LogPayload
		if err := json.Unmarshal([]byte(rawJSON), &payload); err != nil {
			log.Println("Invalid JSON format:", rawJSON)
			continue
		}

		jobs <- payload
	}
}

func processLog(ctx context.Context, workerID int, payload LogPayload, linkRepo repository.LinkRepository, accessLogRepo repository.AccessLogRepository) {
	linkID, err := linkRepo.GetLinkIDByCode(ctx, nil, payload.Code)
	if err != nil {
		log.Printf("[worker-%d] Failed to find link: %s, err=%v\n", workerID, payload.Code, err)
		return
	}

	accessLog := model.AccessLog{
		LinkID:    linkID,
		ShortCode: payload.Code,
		IPAddress: payload.IP,
		UserAgent: payload.UA,
		VisitedAt: time.Unix(payload.TS, 0),
	}

	if err := accessLogRepo.SaveAccessLog(ctx, nil, &accessLog); err != nil {
		log.Printf("[worker-%d] Failed to save log: %s, err=%v\n", workerID, payload.Code, err)
		return
	}

	if os.Getenv("APP_ENV") != "production" {
		log.Printf("[worker-%d] ✅ Saved log for %s\n", workerID, payload.Code)
	}
}
