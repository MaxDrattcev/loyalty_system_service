package repository

import (
	"context"
	"errors"
	"fmt"
	"github.com/MaxDrattcev/loyalty_system_service.git/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrInsufficientFunds = errors.New("insufficient funds")

type userPostgresRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) UserRepository {
	return &userPostgresRepository{
		pool: pool,
	}
}

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

func (u *userPostgresRepository) GetByLogin(ctx context.Context, login string) (models.User, error) {
	query := "SELECT id, login, password FROM users WHERE login = $1"
	var user models.User
	err := WithTxRetry(ctx, u.pool, func(tx pgx.Tx) error {
		err := tx.QueryRow(ctx, query, login).Scan(&user.ID, &user.Login, &user.Password)
		return err
	})
	if err != nil {
		return models.User{}, fmt.Errorf("failed to login: %w", err)
	}
	return user, nil
}

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

func (u *userPostgresRepository) GetByUserID(ctx context.Context, userID int64) (models.User, error) {
	var user models.User
	query := "SELECT id, login, password, current_balance, total_withdrawn, created_at FROM users WHERE id = $1"
	err := WithTxRetry(ctx, u.pool, func(tx pgx.Tx) error {
		err := tx.QueryRow(ctx, query, userID).Scan(&user.ID, &user.Login, &user.Password, &user.CurrentBalance,
			&user.TotalWithdrawn, &user.CreatedAt)
		return err
	})
	if err != nil {
		return models.User{}, fmt.Errorf("failed to find user: %w", err)
	}
	return user, nil
}

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
