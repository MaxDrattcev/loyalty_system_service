package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/MaxDrattcev/loyalty_system_service.git/internal/models"
	"github.com/jackc/pgx/v5"
	pgxconn "github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/require"
)

var _ pgx.Tx = (*fakeTx)(nil)

type fakeTx struct {
	execFn func(ctx context.Context, sql string, args ...any) (pgxconn.CommandTag, error)
}

func (f *fakeTx) Begin(ctx context.Context) (pgx.Tx, error) { panic("not used") }
func (f *fakeTx) Commit(ctx context.Context) error          { panic("not used") }
func (f *fakeTx) Rollback(ctx context.Context) error        { panic("not used") }

func (f *fakeTx) CopyFrom(
	ctx context.Context,
	tableName pgx.Identifier,
	columnNames []string,
	rowSrc pgx.CopyFromSource,
) (int64, error) {
	panic("not used")
}

func (f *fakeTx) SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults {
	panic("not used")
}

func (f *fakeTx) LargeObjects() pgx.LargeObjects { panic("not used") }

func (f *fakeTx) Prepare(ctx context.Context, name, sql string) (*pgxconn.StatementDescription, error) {
	panic("not used")
}

func (f *fakeTx) Exec(ctx context.Context, sql string, args ...any) (pgxconn.CommandTag, error) {
	if f.execFn == nil {
		panic("execFn is nil")
	}
	return f.execFn(ctx, sql, args...)
}

func (f *fakeTx) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	panic("not used")
}

func (f *fakeTx) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	panic("not used")
}

func (f *fakeTx) Conn() *pgx.Conn { panic("not used") }

func TestUserRepository_UpdateTx(t *testing.T) {
	repo := &userPostgresRepository{}
	ctx := context.Background()
	order := models.Order{
		UserID:  10,
		Accrual: 150,
	}

	t.Run("success", func(t *testing.T) {
		tx := &fakeTx{
			execFn: func(ctx context.Context, sql string, args ...any) (pgxconn.CommandTag, error) {
				return pgxconn.NewCommandTag("UPDATE 1"), nil
			},
		}

		err := repo.UpdateTx(ctx, tx, order)
		require.NoError(t, err)
	})

	t.Run("exec error", func(t *testing.T) {
		dbErr := errors.New("db exec failed")
		tx := &fakeTx{
			execFn: func(ctx context.Context, sql string, args ...any) (pgxconn.CommandTag, error) {
				return pgxconn.CommandTag{}, dbErr
			},
		}

		err := repo.UpdateTx(ctx, tx, order)
		require.Error(t, err)
		require.ErrorIs(t, err, dbErr)
		require.Contains(t, err.Error(), "failed to update balance for user")
	})

	t.Run("user not found (rows affected = 0)", func(t *testing.T) {
		tx := &fakeTx{
			execFn: func(ctx context.Context, sql string, args ...any) (pgxconn.CommandTag, error) {
				return pgxconn.NewCommandTag("UPDATE 0"), nil
			},
		}

		err := repo.UpdateTx(ctx, tx, order)
		require.Error(t, err)
		require.Contains(t, err.Error(), "user not found id=10")
	})
}

func TestUserRepository_WithdrawIfEnoughTx(t *testing.T) {
	repo := &userPostgresRepository{}
	ctx := context.Background()

	const userID int64 = 5
	const amount int64 = 250

	t.Run("success", func(t *testing.T) {
		tx := &fakeTx{
			execFn: func(ctx context.Context, sql string, args ...any) (pgxconn.CommandTag, error) {
				return pgxconn.NewCommandTag("UPDATE 1"), nil
			},
		}

		err := repo.WithdrawIfEnoughTx(ctx, tx, userID, amount)
		require.NoError(t, err)
	})

	t.Run("exec error", func(t *testing.T) {
		dbErr := errors.New("db write failed")
		tx := &fakeTx{
			execFn: func(ctx context.Context, sql string, args ...any) (pgxconn.CommandTag, error) {
				return pgxconn.CommandTag{}, dbErr
			},
		}

		err := repo.WithdrawIfEnoughTx(ctx, tx, userID, amount)
		require.Error(t, err)
		require.ErrorIs(t, err, dbErr)
		require.Contains(t, err.Error(), "failed to withdraw for user 5")
	})

	t.Run("insufficient funds (rows affected = 0)", func(t *testing.T) {
		tx := &fakeTx{
			execFn: func(ctx context.Context, sql string, args ...any) (pgxconn.CommandTag, error) {
				return pgxconn.NewCommandTag("UPDATE 0"), nil
			},
		}

		err := repo.WithdrawIfEnoughTx(ctx, tx, userID, amount)
		require.Error(t, err)
		require.ErrorIs(t, err, ErrInsufficientFunds)
		require.Contains(t, err.Error(), "balance for user 5")
	})
}
