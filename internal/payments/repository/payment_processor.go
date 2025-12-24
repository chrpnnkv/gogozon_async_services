package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"HW4/internal/payments/dto"
)

type PaymentProcessor struct {
	db          *sql.DB
	resultTopic string
}

func NewPaymentProcessor(db *sql.DB, resultTopic string) *PaymentProcessor {
	if resultTopic == "" {
		panic("resultTopic is empty")
	}
	return &PaymentProcessor{db: db, resultTopic: resultTopic}
}

func (p *PaymentProcessor) HandlePaymentRequested(ctx context.Context, raw []byte) (bool, error) {
	var req struct {
		MessageID string `json:"message_id"`
		OrderID   string `json:"order_id"`
		UserID    string `json:"user_id"`
		Amount    int64  `json:"amount"`
		CreatedAt string `json:"created_at"`
	}
	if err := json.Unmarshal(raw, &req); err != nil {
		return false, err
	}

	tx, err := p.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return false, err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `INSERT INTO inbox(message_id) VALUES ($1)`, req.MessageID)
	if err != nil {
		return true, tx.Commit()
	}

	var existing string
	err = tx.QueryRowContext(ctx, `SELECT order_id FROM transactions WHERE order_id=$1`, req.OrderID).Scan(&existing)
	if err == nil {
		return true, tx.Commit()
	}
	if err != sql.ErrNoRows {
		return false, err
	}

	res, err := tx.ExecContext(ctx, `
		UPDATE accounts
		SET balance = balance - $1, updated_at = now()
		WHERE user_id = $2 AND balance >= $1
	`, req.Amount, req.UserID)
	if err != nil {
		return false, err
	}
	affected, _ := res.RowsAffected()

	status := "FINISHED"
	reason := ""
	if affected == 0 {
		status = "FAILED"
		reason = "insufficient_funds_or_account_missing"
	} else {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO transactions(order_id, user_id, amount)
			VALUES ($1,$2,$3)
		`, req.OrderID, req.UserID, req.Amount)
		if err != nil {
			return false, err
		}
	}

	ev := dto.PaymentResult{
		MessageID: req.MessageID,
		OrderID:   req.OrderID,
		UserID:    req.UserID,
		Amount:    req.Amount,
		Status:    status,
		Reason:    reason,
		CreatedAt: time.Now().UTC().Format(time.RFC3339Nano),
	}
	payload, _ := json.Marshal(ev)

	_, err = tx.ExecContext(ctx, `
		INSERT INTO outbox(topic, key, payload)
		VALUES ($1,$2,$3)
	`, p.resultTopic, req.OrderID, payload)
	if err != nil {
		return false, err
	}

	return false, tx.Commit()
}
