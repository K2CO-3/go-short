package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"go-short/internal/model"
	"go-short/internal/repository"
	"go-short/internal/repository/impl/postgresql"
	"go-short/internal/repository/impl/redis"

	redisclient "github.com/redis/go-redis/v9"
)

// LogPayload 对应 Redirect Server 发送的 JSON 结构
type LogPayload struct {
	Code string `json:"code"`
	IP   string `json:"ip"`
	UA   string `json:"ua"`
	TS   int64  `json:"ts"`
}

// logJob 包含 payload 与消息 ID，用于消费后 XACK
type logJob struct {
	payload LogPayload
	msgID   string
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

	ctx := context.Background()

	// 创建 Stream 消费组（若不存在）。BUSYGROUP 表示组已存在，可忽略
	if err := rdb.XGroupCreateMkStream(ctx, redis.AccessLogStream, redis.AccessLogConsumerGroup, "0").Err(); err != nil && !strings.Contains(err.Error(), "BUSYGROUP") {
		log.Fatal("Failed to create consumer group:", err)
	}
	log.Printf("👷 Worker started (Redis Stream), pool size=%d, waiting for logs...\n", poolSize)

	jobs := make(chan logJob, poolSize*2)

	// 协程池：多个 worker 并发消费
	for i := 0; i < poolSize; i++ {
		go func(id int) {
			for job := range jobs {
				processLog(ctx, id, job, linkRepo, accessLogRepo, rdb)
			}
		}(i)
	}

	// 主 goroutine：XReadGroup 消费 Stream
	consumerName := "worker"
	for {
		streams, err := rdb.XReadGroup(ctx, &redisclient.XReadGroupArgs{
			Group:    redis.AccessLogConsumerGroup,
			Consumer: consumerName,
			Streams:  []string{redis.AccessLogStream, ">"},
			Count:    10,
			Block:    5 * time.Second,
		}).Result()
		if err != nil {
			if err == redisclient.Nil {
				continue
			}
			log.Println("Redis connect error, retrying in 5s...", err)
			time.Sleep(5 * time.Second)
			continue
		}

		for _, s := range streams {
			for _, msg := range s.Messages {
				raw, ok := msg.Values["payload"].(string)
				if !ok {
					log.Println("Invalid stream message, missing payload")
					_ = rdb.XAck(ctx, redis.AccessLogStream, redis.AccessLogConsumerGroup, msg.ID).Err()
					continue
				}
				var payload LogPayload
				if err := json.Unmarshal([]byte(raw), &payload); err != nil {
					log.Println("Invalid JSON format:", raw)
					_ = rdb.XAck(ctx, redis.AccessLogStream, redis.AccessLogConsumerGroup, msg.ID).Err()
					continue
				}
				jobs <- logJob{payload: payload, msgID: msg.ID}
			}
		}
	}
}

func processLog(ctx context.Context, workerID int, job logJob, linkRepo repository.LinkRepository, accessLogRepo repository.AccessLogRepository, rdb *redisclient.Client) {
	payload := job.payload
	ack := func() { _ = rdb.XAck(ctx, redis.AccessLogStream, redis.AccessLogConsumerGroup, job.msgID).Err() }
	linkID, err := linkRepo.GetLinkIDByCode(ctx, nil, payload.Code)
	if err != nil {
		log.Printf("[worker-%d] Failed to find link: %s, err=%v\n", workerID, payload.Code, err)
		ack() // 链接已不存在，无需重试
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
		// 不 ack，消息留 PENDING，后续可加 XAUTOCLAIM 重试
		return
	}
	ack()

	if os.Getenv("APP_ENV") != "production" {
		log.Printf("[worker-%d] ✅ Saved log for %s\n", workerID, payload.Code)
	}
}
