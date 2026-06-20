package mq

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"go-short/internal/repository"

	"github.com/segmentio/kafka-go"
)

const CacheInvalidateTopic = "cache_invalidate"

// Redirect 事件类型（与 cache_invalidate topic 消息 JSON 字段 event 对应）
const (
	EventCacheInvalidate = "cache_invalidate"
	EventBloomAdd        = "bloom_add"
)

// NewCacheInvalidateWriter 缓存失效 Kafka 生产者（API / Worker 侧写入）
func NewCacheInvalidateWriter() *kafka.Writer {
	return &kafka.Writer{
		Addr:     kafka.TCP(KafkaBrokers()...),
		Topic:    CacheInvalidateTopic,
		Balancer: &kafka.LeastBytes{},
	}
}

// NewCacheInvalidateReader Redirect 侧订阅缓存失效；groupID 须每实例唯一以实现广播到所有 Redirect
func NewCacheInvalidateReader(groupID string) *kafka.Reader {
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers:  KafkaBrokers(),
		Topic:    CacheInvalidateTopic,
		GroupID:  groupID,
		MinBytes: 1,
		MaxBytes: 10e6,
	})
}

// CacheInvalidateConsumerGroupID 默认每个进程唯一，使该 topic 上每条消息被每个 Redirect 副本各消费一次
func CacheInvalidateConsumerGroupID() string {
	if g := os.Getenv("KAFKA_CACHE_INVALIDATE_GROUP"); g != "" {
		return g
	}
	host, err := os.Hostname()
	if err != nil || host == "" {
		host = "unknown"
	}
	return fmt.Sprintf("redirect-cache-invalidate-%s-%d", host, os.Getpid())
}

type kafkaRedirectNotifier struct {
	w *kafka.Writer
}

// NewKafkaRedirectNotifier 通过 Kafka 通知所有 Redirect 删除本地缓存
func NewKafkaRedirectNotifier(w *kafka.Writer) repository.RedirectCacheNotifier {
	return &kafkaRedirectNotifier{w: w}
}

func (k *kafkaRedirectNotifier) NotifyInvalidate(ctx context.Context, shortCode string) error {
	payload, err := json.Marshal(map[string]string{
		"code":  shortCode,
		"event": EventCacheInvalidate,
	})
	if err != nil {
		return err
	}
	return k.w.WriteMessages(ctx, kafka.Message{Value: payload})
}

func (k *kafkaRedirectNotifier) NotifyBloomAdd(ctx context.Context, shortCode string) error {
	payload, err := json.Marshal(map[string]string{
		"code":  shortCode,
		"event": EventBloomAdd,
	})
	if err != nil {
		return err
	}
	return k.w.WriteMessages(ctx, kafka.Message{Value: payload})
}
