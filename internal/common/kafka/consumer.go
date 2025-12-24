package kafka

import (
	"context"
	"strings"
	"time"

	kafkago "github.com/segmentio/kafka-go"
)

type Consumer struct {
	r *kafkago.Reader
}

func NewConsumer(brokers []string, topic, groupID string) *Consumer {
	return &Consumer{
		r: kafkago.NewReader(kafkago.ReaderConfig{
			Brokers:        brokers,
			Topic:          topic,
			GroupID:        groupID,
			MinBytes:       1e3,
			MaxBytes:       10e6,
			CommitInterval: 0,
			StartOffset:    kafkago.FirstOffset,
		}),
	}
}

func (c *Consumer) Fetch(ctx context.Context) (kafkago.Message, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	return c.r.FetchMessage(ctx)
}

func (c *Consumer) Commit(ctx context.Context, msg kafkago.Message) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return c.r.CommitMessages(ctx, msg)
}

func (c *Consumer) Read(ctx context.Context) ([]byte, error) {
	msg, err := c.r.ReadMessage(ctx)
	if err != nil {
		return nil, err
	}
	return msg.Value, nil
}

func (c *Consumer) Close() error {
	return c.r.Close()
}

func SplitBrokers(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
