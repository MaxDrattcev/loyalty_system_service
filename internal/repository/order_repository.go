package repository

import (
	"context"
	"fmt"
	"github.com/MaxDrattcev/loyalty_system_service.git/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// orderRepository implements OrderRepository using PostgreSQL.
type orderRepository struct {
	pool *pgxpool.Pool
}

// NewOrderRepository creates OrderRepository backed by PostgreSQL pool.
func NewOrderRepository(pool *pgxpool.Pool) OrderRepository {
	return &orderRepository{
		pool: pool,
	}
}

// Create inserts order in its own retryable transaction.
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

// CreateTx inserts order using provided transaction.
func (r *orderRepository) CreateTx(ctx context.Context, tx pgx.Tx, order models.Order) error {
	query := "INSERT INTO orders(number, user_id, status) VALUES ($1, $2, $3)"
	if _, err := tx.Exec(ctx, query, order.Number, order.UserID, order.Status); err != nil {
		return fmt.Errorf("failed to create order: %w", err)
	}
	return nil
}

// GetByNumber fetches order by unique order number.
func (r *orderRepository) GetByNumber(ctx context.Context, number int64) (models.Order, error) {
	order, err := queryOneByField(ctx, r.pool, "orders", "id, number, user_id, status, accrual, uploaded_at", "number", number,
		func(row rowScanner) (models.Order, error) {
			order := models.Order{}
			err := row.Scan(&order.ID, &order.Number, &order.UserID, &order.Status, &order.Accrual, &order.UploadedAt)
			return order, err
		})
	if err != nil {
		return models.Order{}, fmt.Errorf("failed to get order by number: %w", err)
	}
	return order, nil
}

// UpdateTx updates order status and accrual using provided transaction.
func (r *orderRepository) UpdateTx(ctx context.Context, tx pgx.Tx, accrual models.Accrual, intAccrual int64) error {
	query := "UPDATE orders SET status = $1, accrual = $2 WHERE number = $3"
	if _, err := tx.Exec(ctx, query, accrual.Status, intAccrual, accrual.OrderNumber); err != nil {
		return fmt.Errorf("failed to update order: %w", err)
	}
	return nil
}

// GetStatusNewAndProcessing returns stale NEW/PROCESSING orders
// that should be rechecked in accrual system.
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

// GetOrdersByUserID returns user orders sorted by upload time descending.
func (r *orderRepository) GetOrdersByUserID(ctx context.Context, userID int64) ([]models.Order, error) {
	orders, err := queryManyByField(ctx, r.pool, "orders", "id, number, user_id, status, accrual, uploaded_at",
		"user_id", userID, "uploaded_at DESC",
		func(row rowScanner) (models.Order, error) {
			var order models.Order
			err := row.Scan(&order.ID, &order.Number, &order.UserID, &order.Status, &order.Accrual, &order.UploadedAt)
			return order, err
		})

	if err != nil {
		return nil, fmt.Errorf("failed to get orders by user id %d: %w", userID, err)
	}
	return orders, nil
}
