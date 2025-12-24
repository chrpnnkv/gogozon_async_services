package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"HW4/internal/orders/dto"
)

type OrdersRepo struct {
	db                    *sql.DB
	paymentRequestedTopic string
}

func NewOrdersRepo(db *sql.DB, paymentRequestedTopic string) *OrdersRepo {
	if paymentRequestedTopic == "" {
		panic(fmt.Errorf("paymentRequestedTopic is empty"))
	}
	return &OrdersRepo{
		db:                    db,
		paymentRequestedTopic: paymentRequestedTopic,
	}
}

type Order struct {
	ID          string
	UserID      string
	Amount      int64
	Description string
	Status      string
	CreatedAt   time.Time
}

func (r *OrdersRepo) CreateOrderWithOutbox(ctx context.Context, userID string, amount int64, description string) (string, error) {
	orderID := uuid.NewString()
	messageID := uuid.NewString()

	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `
		INSERT INTO orders(id, user_id, amount, description, status)
		VALUES ($1,$2,$3,$4,'NEW')
	`, orderID, userID, amount, description)
	if err != nil {
		return "", err
	}

	ev := dto.PaymentRequested{
		MessageID: messageID,
		OrderID:   orderID,
		UserID:    userID,
		Amount:    amount,
		CreatedAt: time.Now().UTC().Format(time.RFC3339Nano),
	}
	payload, _ := json.Marshal(ev)

	_, err = tx.ExecContext(ctx, `
		INSERT INTO outbox(topic, key, payload)
		VALUES ($1,$2,$3)
	`, r.paymentRequestedTopic, orderID, payload)
	if err != nil {
		return "", err
	}

	if err := tx.Commit(); err != nil {
		return "", err
	}
	return orderID, nil
}

func (r *OrdersRepo) ListOrdersByUser(ctx context.Context, userID string) ([]Order, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, amount, description, status, created_at
		FROM orders
		WHERE user_id = $1
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Order
	for rows.Next() {
		var o Order
		if err := rows.Scan(&o.ID, &o.UserID, &o.Amount, &o.Description, &o.Status, &o.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, o)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *OrdersRepo) GetOrderByID(ctx context.Context, id string) (Order, error) {
	var o Order
	err := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, amount, description, status, created_at
		FROM orders
		WHERE id = $1
	`, id).Scan(&o.ID, &o.UserID, &o.Amount, &o.Description, &o.Status, &o.CreatedAt)

	if err == sql.ErrNoRows {
		return Order{}, errors.New("not_found")
	}
	if err != nil {
		return Order{}, err
	}
	return o, nil
}
