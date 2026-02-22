package main

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"go-short/internal/model"
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
	// 1. 初始化数据库连接
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
	accessLogRepo := postgresql.NewAccessLogRepository(db)

	log.Println("👷 Worker started, waiting for logs...")

	ctx := context.Background()

	// 3. 无限循环消费
	for {
		// BLPop: 阻塞式读取列表右侧元素 (Timeout 0 表示无限等待)
		// result[0] 是 key 名, result[1] 是 value
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

		// 4. 根据短码获取链接ID
		linkID, err := linkRepo.GetLinkIDByCode(ctx, nil, payload.Code)
		if err != nil {
			log.Println("Failed to find link:", payload.Code, err)
			continue
		}

		// 5. 构造数据库模型
		accessLog := model.AccessLog{
			LinkID:    linkID,
			ShortCode: payload.Code,
			IPAddress: payload.IP,
			UserAgent: payload.UA,
			VisitedAt: time.Unix(payload.TS, 0),
		}

		// 6. 写入 PostgreSQL
		// 优化思路：高并发下可以使用 Buffer Channel攒一批再 Batch Insert
		if err := accessLogRepo.SaveAccessLog(ctx, nil, &accessLog); err != nil {
			log.Println("Failed to save log to DB:", err)
			// 实际生产中可能需要重新放回 Redis 或死信队列
		} else {
			// 开发环境打印一下，证明 Worker 在工作
			log.Printf("✅ Saved log for %s", payload.Code)
		}
	}
}
