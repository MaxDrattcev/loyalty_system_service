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

// compile-time check
var _ pgx.Tx = (*fakeOrderTx)(nil)

type fakeOrderTx struct {
	execFn func(ctx context.Context, sql string, args ...any) (pgxconn.CommandTag, error)
}

func (f *fakeOrderTx) Begin(ctx context.Context) (pgx.Tx, error) { panic("not used") }
func (f *fakeOrderTx) Commit(ctx context.Context) error          { panic("not used") }
func (f *fakeOrderTx) Rollback(ctx context.Context) error        { panic("not used") }

func (f *fakeOrderTx) CopyFrom(
	ctx context.Context,
	tableName pgx.Identifier,
	columnNames []string,
	rowSrc pgx.CopyFromSource,
) (int64, error) {
	panic("not used")
}

func (f *fakeOrderTx) SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults {
	panic("not used")
}

func (f *fakeOrderTx) LargeObjects() pgx.LargeObjects { panic("not used") }

func (f *fakeOrderTx) Prepare(ctx context.Context, name, sql string) (*pgxconn.StatementDescription, error) {
	panic("not used")
}

func (f *fakeOrderTx) Exec(ctx context.Context, sql string, args ...any) (pgxconn.CommandTag, error) {
	if f.execFn == nil {
		panic("execFn is nil")
	}
	return f.execFn(ctx, sql, args...)
}

func (f *fakeOrderTx) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	panic("not used")
}

func (f *fakeOrderTx) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	panic("not used")
}

func (f *fakeOrderTx) Conn() *pgx.Conn { panic("not used") }

func TestOrderRepository_CreateTx(t *testing.T) {
	repo := &orderRepository{}
	ctx := context.Background()

	order := models.Order{
		Number: 79927398713,
		UserID: 10,
		Status: models.OrderStatusNew,
	}

	t.Run("success", func(t *testing.T) {
		tx := &fakeTx{
			execFn: func(ctx context.Context, sql string, args ...any) (pgxconn.CommandTag, error) {
				require.Equal(t, "INSERT INTO orders(number, user_id, status) VALUES ($1, $2, $3)", sql)
				require.Equal(t, 3, len(args))
				require.Equal(t, order.Number, args[0])
				require.Equal(t, order.UserID, args[1])
				require.Equal(t, order.Status, args[2])

				return pgxconn.NewCommandTag("INSERT 0 1"), nil
			},
		}

		err := repo.CreateTx(ctx, tx, order)
		require.NoError(t, err)
	})

	t.Run("exec error", func(t *testing.T) {
		dbErr := errors.New("insert failed")
		tx := &fakeTx{
			execFn: func(ctx context.Context, sql string, args ...any) (pgxconn.CommandTag, error) {
				return pgxconn.CommandTag{}, dbErr
			},
		}

		err := repo.CreateTx(ctx, tx, order)
		require.Error(t, err)
		require.ErrorIs(t, err, dbErr)
		require.Contains(t, err.Error(), "failed to create order")
	})
}

func TestOrderRepository_UpdateTx(t *testing.T) {
	repo := &orderRepository{}
	ctx := context.Background()

	accrual := models.Accrual{
		OrderNumber: "79927398713",
		Status:      string(models.OrderStatusProcessed),
		Accrual:     12.34,
	}
	intAccrual := int64(1234)

	t.Run("success", func(t *testing.T) {
		tx := &fakeTx{
			execFn: func(ctx context.Context, sql string, args ...any) (pgxconn.CommandTag, error) {
				require.Equal(t, "UPDATE orders SET status = $1, accrual = $2 WHERE number = $3", sql)
				require.Equal(t, 3, len(args))
				require.Equal(t, accrual.Status, args[0])
				require.Equal(t, intAccrual, args[1])
				require.Equal(t, accrual.OrderNumber, args[2])

				return pgxconn.NewCommandTag("UPDATE 1"), nil
			},
		}

		err := repo.UpdateTx(ctx, tx, accrual, intAccrual)
		require.NoError(t, err)
	})

	t.Run("exec error", func(t *testing.T) {
		dbErr := errors.New("update failed")
		tx := &fakeTx{
			execFn: func(ctx context.Context, sql string, args ...any) (pgxconn.CommandTag, error) {
				return pgxconn.CommandTag{}, dbErr
			},
		}

		err := repo.UpdateTx(ctx, tx, accrual, intAccrual)
		require.Error(t, err)
		require.ErrorIs(t, err, dbErr)
		require.Contains(t, err.Error(), "failed to update order")
	})
}
