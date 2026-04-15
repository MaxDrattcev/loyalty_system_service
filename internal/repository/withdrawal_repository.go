package repository

import (
	"context"
	"fmt"
	"github.com/MaxDrattcev/loyalty_system_service.git/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// withdrawalRepository implements WithdrawalRepository using PostgreSQL.
type withdrawalRepository struct {
	pool *pgxpool.Pool
}

// NewWithdrawalRepository creates WithdrawalRepository backed by PostgreSQL pool.
func NewWithdrawalRepository(pool *pgxpool.Pool) WithdrawalRepository {
	return &withdrawalRepository{
		pool: pool,
	}
}

// CreateTx inserts withdrawal history using provided transaction.
func (r *withdrawalRepository) CreateTx(ctx context.Context, tx pgx.Tx, wh models.WithdrawalHistory) error {
	query := "INSERT INTO withdrawal_history (user_id, order_number, sum_withdrawal) VALUES ($1, $2, $3)"
	if _, err := tx.Exec(ctx, query, wh.UserID, wh.OrderNumber, wh.SumWithdrawal); err != nil {
		return fmt.Errorf("failed create withdrawal history: %w", err)
	}
	return nil
}

// GetWithdrawals fetches withdrawal history for user sorted by processed time descending.
func (r *withdrawalRepository) GetWithdrawals(ctx context.Context, userID int64) ([]models.WithdrawalHistory, error) {
	withdrawals, err := queryManyByField(ctx, r.pool, "withdrawal_history",
		"id, user_id, order_number, sum_withdrawal, processed_at", "user_id", userID, "processed_at DESC",
		func(row rowScanner) (models.WithdrawalHistory, error) {
			var w models.WithdrawalHistory
			err := row.Scan(&w.ID, &w.UserID, &w.OrderNumber, &w.SumWithdrawal, &w.ProcessedAt)
			return w, err
		})
	if err != nil {
		return nil, fmt.Errorf("failed get withdrawals for user %d: %w", userID, err)
	}
	return withdrawals, nil
}
