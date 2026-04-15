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
var _ pgx.Tx = (*fakeWithdrawalTx)(nil)

type fakeWithdrawalTx struct {
	execFn func(ctx context.Context, sql string, args ...any) (pgxconn.CommandTag, error)
}

func (f *fakeWithdrawalTx) Begin(ctx context.Context) (pgx.Tx, error) { panic("not used") }
func (f *fakeWithdrawalTx) Commit(ctx context.Context) error          { panic("not used") }
func (f *fakeWithdrawalTx) Rollback(ctx context.Context) error        { panic("not used") }

func (f *fakeWithdrawalTx) CopyFrom(
	ctx context.Context,
	tableName pgx.Identifier,
	columnNames []string,
	rowSrc pgx.CopyFromSource,
) (int64, error) {
	panic("not used")
}

func (f *fakeWithdrawalTx) SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults {
	panic("not used")
}

func (f *fakeWithdrawalTx) LargeObjects() pgx.LargeObjects { panic("not used") }

func (f *fakeWithdrawalTx) Prepare(ctx context.Context, name, sql string) (*pgxconn.StatementDescription, error) {
	panic("not used")
}

func (f *fakeWithdrawalTx) Exec(ctx context.Context, sql string, args ...any) (pgxconn.CommandTag, error) {
	if f.execFn == nil {
		panic("execFn is nil")
	}
	return f.execFn(ctx, sql, args...)
}

func (f *fakeWithdrawalTx) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	panic("not used")
}

func (f *fakeWithdrawalTx) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	panic("not used")
}

func (f *fakeWithdrawalTx) Conn() *pgx.Conn { panic("not used") }

func TestWithdrawalRepository_CreateTx(t *testing.T) {
	repo := &withdrawalRepository{}
	ctx := context.Background()

	wh := models.WithdrawalHistory{
		UserID:        10,
		OrderNumber:   79927398713,
		SumWithdrawal: 1234,
	}

	t.Run("success", func(t *testing.T) {
		tx := &fakeWithdrawalTx{
			execFn: func(ctx context.Context, sql string, args ...any) (pgxconn.CommandTag, error) {
				require.Equal(t, "INSERT INTO withdrawal_history (user_id, order_number, sum_withdrawal) VALUES ($1, $2, $3)", sql)
				require.Len(t, args, 3)
				require.Equal(t, wh.UserID, args[0])
				require.Equal(t, wh.OrderNumber, args[1])
				require.Equal(t, wh.SumWithdrawal, args[2])

				return pgxconn.NewCommandTag("INSERT 0 1"), nil
			},
		}

		err := repo.CreateTx(ctx, tx, wh)
		require.NoError(t, err)
	})

	t.Run("exec error", func(t *testing.T) {
		dbErr := errors.New("insert failed")
		tx := &fakeWithdrawalTx{
			execFn: func(ctx context.Context, sql string, args ...any) (pgxconn.CommandTag, error) {
				return pgxconn.CommandTag{}, dbErr
			},
		}

		err := repo.CreateTx(ctx, tx, wh)
		require.Error(t, err)
		require.ErrorIs(t, err, dbErr)
		require.Contains(t, err.Error(), "failed create withdrawal history")
	})
}
