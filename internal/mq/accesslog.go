package mq

import (
	"os"
	"strings"

	"github.com/segmentio/kafka-go"
)

const (
	AccessLogTopic = "access_logs"
)

// KafkaBrokers 从环境变量读取 Kafka 地址，默认 localhost:9092
func KafkaBrokers() []string {
	s := os.Getenv("KAFKA_BROKERS")
	if s == "" {
		return []string{"localhost:9092"}
	}
	return strings.Split(s, ",")
}

// NewAccessLogWriter 创建访问日志 Kafka 生产者
func NewAccessLogWriter() *kafka.Writer {
	return &kafka.Writer{
		Addr:     kafka.TCP(KafkaBrokers()...),
		Topic:    AccessLogTopic,
		Balancer: &kafka.LeastBytes{},
	}
}

// NewAccessLogReader 创建访问日志 Kafka 消费者（消费者组）
func NewAccessLogReader(groupID string) *kafka.Reader {
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers:  KafkaBrokers(),
		Topic:    AccessLogTopic,
		GroupID:  groupID,
		MinBytes: 1,
		MaxBytes: 10e6,
	})
}
