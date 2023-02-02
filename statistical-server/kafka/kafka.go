package kafka

import (
	"time"

	"coin-server/common/statistical2/models"
	"coin-server/statistical-server/config"

	"github.com/segmentio/kafka-go"
)

var Reader *kafka.Reader
var LoginReader *kafka.Reader

func Init() {
	cfg := kafka.ReaderConfig{
		Brokers:        config.Kafka.Addr,
		GroupID:        "statistical",
		GroupTopics:    models.TopicList, // 消费组关注的多个topic
		CommitInterval: 10 * time.Millisecond,
	}
	if config.Kafka.MinBytes > 0 {
		cfg.MinBytes = config.Kafka.MinBytes
	}
	if config.Kafka.MaxBytes > 0 {
		cfg.MaxBytes = config.Kafka.MaxBytes
	}
	Reader = kafka.NewReader(cfg)
}

func InitLogin() {
	cfg := kafka.ReaderConfig{
		Brokers:        config.Kafka.Addr,
		GroupID:        "login_statistical",
		GroupTopics:    models.LoginTopicList, // 消费组关注的多个topic
		CommitInterval: 10 * time.Millisecond,
	}
	if config.Kafka.MinBytes > 0 {
		cfg.MinBytes = config.Kafka.MinBytes
	}
	if config.Kafka.MaxBytes > 0 {
		cfg.MaxBytes = config.Kafka.MaxBytes
	}
	LoginReader = kafka.NewReader(cfg)
}
