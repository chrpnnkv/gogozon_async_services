package worker

import (
	"context"
	"encoding/json"
	"log"

	"HW4/internal/common/kafka"
	"HW4/internal/orders/dto"
	"HW4/internal/orders/repository"
)

type PaymentResultConsumer struct {
	consumer *kafka.Consumer
	repo     *repository.OrdersStatusRepo
}

func NewPaymentResultConsumer(consumer *kafka.Consumer, repo *repository.OrdersStatusRepo) *PaymentResultConsumer {
	return &PaymentResultConsumer{consumer: consumer, repo: repo}
}

func (c *PaymentResultConsumer) Run(ctx context.Context) {
	for {
		msg, err := c.consumer.Fetch(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			log.Printf("[orders-consumer] fetch error: %v", err)
			continue
		}

		log.Printf("[orders-consumer] got msg topic=%s partition=%d offset=%d key=%s value=%s",
			msg.Topic, msg.Partition, msg.Offset, string(msg.Key), string(msg.Value))

		var ev dto.PaymentResult
		if err := json.Unmarshal(msg.Value, &ev); err != nil {
			log.Printf("[orders-consumer] bad json: %v; value=%s", err, string(msg.Value))
			_ = c.consumer.Commit(ctx, msg) // чтобы не зациклиться
			continue
		}

		newStatus := "FAILED"
		if ev.Status == "FINISHED" {
			newStatus = "FINISHED"
		}

		updated, err := c.repo.SetStatusIfNew(ctx, ev.OrderID, newStatus)
		if err != nil {
			log.Printf("[orders-consumer] db error: %v", err)
			continue
		}

		if err := c.consumer.Commit(ctx, msg); err != nil {
			log.Printf("[orders-consumer] commit error: %v", err)
			continue
		}

		log.Printf("[orders-consumer] order=%s eventStatus=%s -> set=%s updated=%v",
			ev.OrderID, ev.Status, newStatus, updated)
	}
}
