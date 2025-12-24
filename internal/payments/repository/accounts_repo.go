package repository

import (
	"context"
	"database/sql"

	"github.com/lib/pq"
)

type AccountsRepo struct {
	db *sql.DB
}

func NewAccountsRepo(db *sql.DB) *AccountsRepo { return &AccountsRepo{db: db} }

func (r *AccountsRepo) Create(ctx context.Context, userID string, balance int64) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO accounts(user_id, balance) VALUES ($1,$2)
	`, userID, balance)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return err
		}
	}
	return err
}

func (r *AccountsRepo) TopUp(ctx context.Context, userID string, amount int64) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE accounts SET balance = balance + $1, updated_at = now()
		WHERE user_id = $2
	`, amount, userID)
	return err
}

func (r *AccountsRepo) GetBalance(ctx context.Context, userID string) (int64, error) {
	var b int64
	err := r.db.QueryRowContext(ctx, `SELECT balance FROM accounts WHERE user_id=$1`, userID).Scan(&b)
	return b, err
}
