package repository

import (
	"context"
	"fmt"
	"github.com/MaxDrattcev/loyalty_system_service.git/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type withdrawalRepository struct {
	pool *pgxpool.Pool
}

func NewWithdrawalRepository(pool *pgxpool.Pool) WithdrawalRepository {
	return &withdrawalRepository{
		pool: pool,
	}
}

func (r *withdrawalRepository) CreateTx(ctx context.Context, tx pgx.Tx, wh models.WithdrawalHistory) error {
	query := "INSERT INTO withdrawal_history (user_id, order_number, sum_withdrawal) VALUES ($1, $2, $3)"
	if _, err := tx.Exec(ctx, query, wh.UserID, wh.OrderNumber, wh.SumWithdrawal); err != nil {
		return fmt.Errorf("failed create withdrawal history: %w", err)
	}
	return nil
}

func (r *withdrawalRepository) GetWithdrawals(ctx context.Context, userID int64) ([]models.WithdrawalHistory, error) {
	var withdrawals []models.WithdrawalHistory
	query := "SELECT id, user_id, order_number, sum_withdrawal, processed_at " +
		"FROM withdrawal_history WHERE user_id = $1 " +
		"ORDER BY processed_at DESC"

	err := WithTxRetry(ctx, r.pool, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx, query, userID)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var withdrawal models.WithdrawalHistory
			if err = rows.Scan(&withdrawal.ID, &withdrawal.UserID, &withdrawal.OrderNumber, &withdrawal.SumWithdrawal,
				&withdrawal.ProcessedAt); err != nil {
				return err
			}
			withdrawals = append(withdrawals, withdrawal)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, fmt.Errorf("failed get withdrawals for user %d: %w", userID, err)
	}
	return withdrawals, nil
}
