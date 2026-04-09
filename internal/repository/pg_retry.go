package repository

import (
	"context"
	"errors"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"strings"
	"time"
)

var retryablePGCodes = map[string]bool{
	"40000": true,
	"40001": true,
	"40003": true,
	"40P01": true,

	"53000": true,
	"53100": true,
	"53200": true,
	"53300": true,

	"57000": true,
	"57P01": true,
	"57P02": true,
	"57P03": true,

	"55P03": true,

	"58030": true,
}

var dbRetryDelays = []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second}

type Tx = pgx.Tx

func WithTxRetry(ctx context.Context, pool *pgxpool.Pool, fn func(tx pgx.Tx) error) error {
	var lastErr error

	for attempt := 0; ; attempt++ {
		tx, err := pool.Begin(ctx)
		if err != nil {
			if !isRetryablePGTransportError(err) {
				return err
			}
			lastErr = err
			if attempt >= len(dbRetryDelays) {
				return lastErr
			}
			if err := sleep(ctx, dbRetryDelays[attempt]); err != nil {
				return err
			}
			continue
		}
		err = fn(tx)
		if err == nil {
			return tx.Commit(ctx)
		}
		_ = tx.Rollback(ctx)
		if !isRetryablePGTransportError(err) {
			return err
		}
		lastErr = err
		if attempt >= len(dbRetryDelays) {
			return lastErr
		}
		if err := sleep(ctx, dbRetryDelays[attempt]); err != nil {
			return err
		}
	}
}

func isRetryablePGTransportError(err error) bool {
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}

	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}

	if strings.HasPrefix(pgErr.Code, "08") {
		return true
	}
	return retryablePGCodes[pgErr.Code]
}

func sleep(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}
