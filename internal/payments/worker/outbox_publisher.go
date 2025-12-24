package worker

import (
	"context"
	"database/sql"
	"log"
	"time"

	"HW4/internal/common/kafka"
)

type OutboxPublisher struct {
	db        *sql.DB
	producer  *kafka.Producer
	batchSize int
}

type outboxRow struct {
	ID      int64
	Topic   string
	Key     string
	Payload []byte
}

func NewOutboxPublisher(db *sql.DB, producer *kafka.Producer) *OutboxPublisher {
	return &OutboxPublisher{
		db:        db,
		producer:  producer,
		batchSize: 20,
	}
}

func (w *OutboxPublisher) Run(ctx context.Context) {
	ticker := time.NewTicker(700 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := w.tick(ctx); err != nil {
				log.Printf("[outbox] tick error: %v", err)
			}
		}
	}
}

func (w *OutboxPublisher) tick(ctx context.Context) error {
	batch, err := w.lockAndFetchBatch(ctx)
	if err != nil {
		return err
	}
	if len(batch) == 0 {
		return nil
	}

	for _, r := range batch {
		err := w.producer.Publish(ctx, r.Topic, []byte(r.Key), r.Payload)
		if err != nil {
			_ = w.fail(ctx, r.ID, err)
			continue
		}
		_ = w.success(ctx, r.ID)
	}
	return nil
}

func (w *OutboxPublisher) lockAndFetchBatch(ctx context.Context) ([]outboxRow, error) {
	tx, err := w.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Постгрес-фокус: UPDATE + SKIP LOCKED + RETURNING
	rows, err := tx.QueryContext(ctx, `
		WITH picked AS (
			SELECT id
			FROM outbox
			WHERE processed_at IS NULL
			  AND next_retry_at <= now()
			  AND (locked_until IS NULL OR locked_until < now())
			ORDER BY id
			FOR UPDATE SKIP LOCKED
			LIMIT $1
		)
		UPDATE outbox o
		SET locked_until = now() + interval '10 seconds'
		FROM picked
		WHERE o.id = picked.id
		RETURNING o.id, o.topic, o.key, o.payload
	`, w.batchSize)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var batch []outboxRow
	for rows.Next() {
		var r outboxRow
		if err := rows.Scan(&r.ID, &r.Topic, &r.Key, &r.Payload); err != nil {
			return nil, err
		}
		batch = append(batch, r)
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return batch, nil
}

func (w *OutboxPublisher) success(ctx context.Context, id int64) error {
	_, err := w.db.ExecContext(ctx, `
		UPDATE outbox
		SET processed_at = now(),
		    locked_until = NULL,
		    last_error = NULL
		WHERE id = $1
	`, id)
	return err
}

func (w *OutboxPublisher) fail(ctx context.Context, id int64, cause error) error {
	_, err := w.db.ExecContext(ctx, `
		UPDATE outbox
		SET attempts = attempts + 1,
		    next_retry_at = now() + (attempts + 1) * interval '2 seconds',
		    last_error = $2,
		    locked_until = NULL
		WHERE id = $1
	`, id, cause.Error())
	return err
}
