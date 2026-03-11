package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"go-short/internal/model"
	"go-short/internal/mq"
	"go-short/internal/repository"
	"go-short/internal/repository/impl/postgresql"
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

	linkRepo := postgresql.NewLinkRepository(db)
	accessLogRepo := postgresql.NewAccessLogRepository(db)

	ctx := context.Background()

	reader := mq.NewAccessLogReader("access_logs_group")
	defer reader.Close()

	log.Printf("👷 Worker started (Kafka), waiting for logs...\n")

	// 从 Kafka 消费消息，仅处理成功后提交 offset（失败则重试）
	for {
		msg, err := reader.FetchMessage(ctx)
		if err != nil {
			log.Println("Kafka read error, retrying in 5s...", err)
			time.Sleep(5 * time.Second)
			continue
		}
		if processLog(ctx, 0, msg.Value, linkRepo, accessLogRepo) {
			if err := reader.CommitMessages(ctx, msg); err != nil {
				log.Println("Kafka commit error:", err)
			}
		}
	}
}

// processLog 处理单条访问日志，返回 true 表示成功（可提交 offset）
func processLog(ctx context.Context, workerID int, data []byte, linkRepo repository.LinkRepository, accessLogRepo repository.AccessLogRepository) bool {
	var payload LogPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		log.Println("Invalid JSON format:", string(data))
		return true // 格式错误无需重试
	}

	linkID, err := linkRepo.GetLinkIDByCode(ctx, nil, payload.Code)
	if err != nil {
		log.Printf("[worker-%d] Failed to find link: %s, err=%v\n", workerID, payload.Code, err)
		return true // 链接已删除，无需重试
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
		return false // DB 失败，不提交，稍后重试
	}

	if os.Getenv("APP_ENV") != "production" {
		log.Printf("[worker-%d] ✅ Saved log for %s\n", workerID, payload.Code)
	}
	return true
}
