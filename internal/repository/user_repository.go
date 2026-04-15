package repository

import (
	"context"
	"errors"
	"fmt"
	"github.com/MaxDrattcev/loyalty_system_service.git/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrInsufficientFunds indicates that user balance is not enough for withdrawal.
var ErrInsufficientFunds = errors.New("insufficient funds")

// userPostgresRepository implements UserRepository using PostgreSQL.
type userPostgresRepository struct {
	pool *pgxpool.Pool
}

// NewUserRepository creates UserRepository backed by PostgreSQL pool.
func NewUserRepository(pool *pgxpool.Pool) UserRepository {
	return &userPostgresRepository{
		pool: pool,
	}
}

// Create inserts user into storage and returns created user with generated ID.
func (u *userPostgresRepository) Create(ctx context.Context, user models.User) (models.User, error) {
	query := "INSERT INTO users (login, password) VALUES ($1, $2) RETURNING id"

	err := WithTxRetry(ctx, u.pool, func(tx pgx.Tx) error {
		err := tx.QueryRow(ctx, query, user.Login, user.Password).Scan(&user.ID)
		return err
	})
	if err != nil {
		return models.User{}, fmt.Errorf("failed to insert user: %w", err)
	}
	return user, nil
}

// GetByLogin fetches user by login.
func (u *userPostgresRepository) GetByLogin(ctx context.Context, login string) (models.User, error) {
	user, err := queryOneByField(ctx, u.pool, "users", "id, login, password", "login", login,
		func(row rowScanner) (models.User, error) {
			var user models.User
			err := row.Scan(&user.ID, &user.Login, &user.Password)
			return user, err
		})

	if err != nil {
		return models.User{}, fmt.Errorf("failed to login: %w", err)
	}
	return user, nil
}

// UpdateTx adds order accrual to user balance inside transaction.
func (u *userPostgresRepository) UpdateTx(ctx context.Context, tx pgx.Tx, order models.Order) error {
	query := "UPDATE users SET current_balance = current_balance + $1 WHERE id = $2"
	tag, err := tx.Exec(ctx, query, order.Accrual, order.UserID)
	if err != nil {
		return fmt.Errorf("failed to update balance for user: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("user not found id=%d", order.UserID)
	}
	return nil
}

// Update adds order accrual to user balance in its own retryable transaction.
func (u *userPostgresRepository) Update(ctx context.Context, order models.Order) error {
	query := "UPDATE users SET current_balance = current_balance + $1 " +
		"WHERE id = $2"
	err := WithTxRetry(ctx, u.pool, func(tx pgx.Tx) error {
		tag, err := tx.Exec(ctx, query, order.Accrual, order.UserID)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return fmt.Errorf("user not found id=%d", order.UserID)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to update balance for user: %w", err)
	}
	return nil
}

// GetByUserID fetches user by ID.
func (u *userPostgresRepository) GetByUserID(ctx context.Context, userID int64) (models.User, error) {
	user, err := queryOneByField(ctx, u.pool, "users",
		"id, login, password, current_balance, total_withdrawn, created_at", "id", userID,
		func(row rowScanner) (models.User, error) {
			var user models.User
			err := row.Scan(&user.ID, &user.Login, &user.Password, &user.CurrentBalance, &user.TotalWithdrawn, &user.CreatedAt)
			return user, err
		})
	if err != nil {
		return models.User{}, fmt.Errorf("failed to find user: %w", err)
	}
	return user, nil
}

// WithdrawIfEnoughTx withdraws amount and increases total withdrawn inside transaction.
// Returns ErrInsufficientFunds when rows are not updated due to insufficient balance.
func (u *userPostgresRepository) WithdrawIfEnoughTx(ctx context.Context, tx pgx.Tx, userID int64, amount int64) error {
	query := `
		UPDATE users
		SET current_balance = current_balance - $1,
		    total_withdrawn = total_withdrawn + $1
		WHERE id = $2
		  AND current_balance >= $1
	`
	tag, err := tx.Exec(ctx, query, amount, userID)
	if err != nil {
		return fmt.Errorf("failed to withdraw for user %d: %w", userID, err)
	}
	if tag.RowsAffected() == 0 {

		return fmt.Errorf("balance for user %d: %w", userID, ErrInsufficientFunds)
	}
	return nil
}
