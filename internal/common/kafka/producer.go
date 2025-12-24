package kafka

import (
	"context"
	"time"

	kafkago "github.com/segmentio/kafka-go"
)

type Producer struct {
	w *kafkago.Writer
}

func NewProducer(brokers []string) *Producer {
	return &Producer{
		w: &kafkago.Writer{
			Addr:         kafkago.TCP(brokers...),
			Balancer:     &kafkago.Hash{},
			RequiredAcks: kafkago.RequireAll,
			Async:        false,
		},
	}
}

func (p *Producer) Close() error { return p.w.Close() }

func (p *Producer) Publish(ctx context.Context, topic string, key []byte, value []byte) error {
	return p.PublishWithHeaders(ctx, topic, key, value, nil)
}

func (p *Producer) PublishWithHeaders(ctx context.Context, topic string, key []byte, value []byte, headers []kafkago.Header) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return p.w.WriteMessages(ctx, kafkago.Message{
		Topic:   topic,
		Key:     key,
		Value:   value,
		Time:    time.Now(),
		Headers: headers,
	})
}
