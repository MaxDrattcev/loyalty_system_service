package repository

import (
	"context"
	"fmt"
	"github.com/MaxDrattcev/loyalty_system_service.git/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type orderRepository struct {
	pool *pgxpool.Pool
}

func NewOrderRepository(pool *pgxpool.Pool) OrderRepository {
	return &orderRepository{
		pool: pool,
	}
}

func (r *orderRepository) Create(ctx context.Context, order models.Order) error {
	query := "INSERT INTO orders(number, user_id, status) VALUES ($1, $2, $3)"

	err := WithTxRetry(ctx, r.pool, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, query, order.Number, order.UserID, order.Status); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to create order: %w", err)
	}
	return nil
}

func (r *orderRepository) CreateTx(ctx context.Context, tx pgx.Tx, order models.Order) error {
	query := "INSERT INTO orders(number, user_id, status) VALUES ($1, $2, $3)"
	if _, err := tx.Exec(ctx, query, order.Number, order.UserID, order.Status); err != nil {
		return fmt.Errorf("failed to create order: %w", err)
	}
	return nil
}

func (r *orderRepository) GetByNumber(ctx context.Context, number int64) (models.Order, error) {
	query := "SELECT id, number, user_id, status, accrual, uploaded_at FROM orders WHERE number = $1"
	var order models.Order
	err := WithTxRetry(ctx, r.pool, func(tx pgx.Tx) error {
		row := tx.QueryRow(ctx, query, number)
		return row.Scan(&order.ID, &order.Number, &order.UserID, &order.Status, &order.Accrual, &order.UploadedAt)
	})
	if err != nil {
		return models.Order{}, fmt.Errorf("failed to get order by number: %w", err)
	}
	return order, nil
}

func (r *orderRepository) UpdateTx(ctx context.Context, tx pgx.Tx, accrual models.Accrual, intAccrual int64) error {
	query := "UPDATE orders SET status = $1, accrual = $2 WHERE number = $3"
	if _, err := tx.Exec(ctx, query, accrual.Status, intAccrual, accrual.OrderNumber); err != nil {
		return fmt.Errorf("failed to update order: %w", err)
	}
	return nil
}

func (r *orderRepository) Update(ctx context.Context, accrual models.Accrual, intAccrual int64) error {
	query := "UPDATE orders SET status = $1, accrual = $2 WHERE number = $3"

	err := WithTxRetry(ctx, r.pool, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, query, accrual.Status, intAccrual, accrual.OrderNumber); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to update order: %w", err)
	}
	return nil
}

func (r *orderRepository) GetStatusNewAndProcessing(ctx context.Context) ([]models.Order, error) {
	var orders []models.Order
	query := "SELECT id, number, user_id, status, accrual, uploaded_at " +
		"FROM orders WHERE (status = 'NEW' OR status = 'PROCESSING') " +
		"AND uploaded_at < NOW() - INTERVAL '4 minutes'"
	err := WithTxRetry(ctx, r.pool, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx, query)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var order models.Order
			if err := rows.Scan(&order.ID, &order.Number, &order.UserID, &order.Status, &order.Accrual, &order.UploadedAt); err != nil {
				return err
			}
			orders = append(orders, order)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get orders status: new and processing: %w", err)
	}
	return orders, nil
}

func (r *orderRepository) GetOrdersByUserID(ctx context.Context, userID int64) ([]models.Order, error) {
	var orders []models.Order
	query := "SELECT id, number, user_id, status, accrual, uploaded_at " +
		"FROM orders WHERE user_id = $1 " +
		"ORDER BY uploaded_at DESC"
	err := WithTxRetry(ctx, r.pool, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx, query, userID)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var order models.Order
			if err := rows.Scan(&order.ID, &order.Number, &order.UserID, &order.Status, &order.Accrual, &order.UploadedAt); err != nil {
				return err
			}
			orders = append(orders, order)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get orders by user id %d: %w", userID, err)
	}
	return orders, nil
}
