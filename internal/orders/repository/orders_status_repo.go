package repository

import (
	"context"
	"database/sql"
)

type OrdersStatusRepo struct {
	db *sql.DB
}

func NewOrdersStatusRepo(db *sql.DB) *OrdersStatusRepo { return &OrdersStatusRepo{db: db} }

func (r *OrdersStatusRepo) SetStatusIfNew(ctx context.Context, orderID, status string) (bool, error) {
	res, err := r.db.ExecContext(ctx, `
		UPDATE orders
		SET status = $2,
	    updated_at = now()
		WHERE id = $1 AND status = 'NEW'
	`, orderID, status)
	if err != nil {
		return false, err
	}
	aff, _ := res.RowsAffected()
	return aff > 0, nil
}
