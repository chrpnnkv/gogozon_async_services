package worker

import (
	"context"
	"log"
	"os"
	"strconv"

	kafkago "github.com/segmentio/kafka-go"

	"HW4/internal/common/kafka"
	"HW4/internal/payments/repository"
)

type PaymentRequestedConsumer struct {
	consumer   *kafka.Consumer
	processor  *repository.PaymentProcessor
	producer   *kafka.Producer
	retryTopic string
	dlqTopic   string
	retryMax   int
}

func NewPaymentRequestedConsumer(consumer *kafka.Consumer, processor *repository.PaymentProcessor, producer *kafka.Producer) *PaymentRequestedConsumer {
	retryMax := mustIntEnv("KAFKA_RETRY_MAX")
	retryTopic := mustEnv("KAFKA_RETRY_TOPIC")
	dlqTopic := mustEnv("KAFKA_DLQ_TOPIC")

	return &PaymentRequestedConsumer{
		consumer:   consumer,
		processor:  processor,
		producer:   producer,
		retryTopic: retryTopic,
		dlqTopic:   dlqTopic,
		retryMax:   retryMax,
	}
}

func (c *PaymentRequestedConsumer) Run(ctx context.Context) {
	for {
		msg, err := c.consumer.Fetch(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			log.Printf("[payments-consumer] fetch error: %v", err)
			continue
		}

		already, err := c.processor.HandlePaymentRequested(ctx, msg.Value)
		if err != nil {
			c.handleFailure(ctx, msg, err)
			continue
		}

		if err := c.consumer.Commit(ctx, msg); err != nil {
			log.Printf("[payments-consumer] commit error: %v", err)
			continue
		}

		if already {
			log.Printf("[payments-consumer] duplicate ignored")
		} else {
			log.Printf("[payments-consumer] processed order=%s", string(msg.Key))
		}
	}
}

func (c *PaymentRequestedConsumer) handleFailure(ctx context.Context, msg kafkago.Message, cause error) {
	retryCount := kafka.GetRetryCount(msg)

	if retryCount < c.retryMax {
		newHeaders := kafka.WithRetryCount(msg.Headers, retryCount+1)
		pubErr := c.producer.PublishWithHeaders(ctx, c.retryTopic, msg.Key, msg.Value, newHeaders)
		if pubErr != nil {
			log.Printf("[payments-consumer] retry publish failed: %v (orig err=%v)", pubErr, cause)
			return
		}

		if err := c.consumer.Commit(ctx, msg); err != nil {
			log.Printf("[payments-consumer] commit after retry publish failed: %v", err)
			return
		}

		log.Printf("[payments-consumer] moved to retry topic=%s retry_count=%d cause=%v", c.retryTopic, retryCount+1, cause)
		return
	}

	newHeaders := kafka.WithRetryCount(msg.Headers, retryCount)
	pubErr := c.producer.PublishWithHeaders(ctx, c.dlqTopic, msg.Key, msg.Value, newHeaders)
	if pubErr != nil {
		log.Printf("[payments-consumer] dlq publish failed: %v (orig err=%v)", pubErr, cause)
		return
	}

	if err := c.consumer.Commit(ctx, msg); err != nil {
		log.Printf("[payments-consumer] commit after dlq publish failed: %v", err)
		return
	}

	log.Printf("[payments-consumer] moved to DLQ topic=%s retry_count=%d cause=%v", c.dlqTopic, retryCount, cause)
}

func mustEnv(k string) string {
	v := os.Getenv(k)
	if v == "" {
		panic("missing env " + k)
	}
	return v
}

func mustIntEnv(k string) int {
	raw := mustEnv(k)
	n, err := strconv.Atoi(raw)
	if err != nil || n < 0 {
		panic("bad int env " + k + "=" + raw)
	}
	return n
}
