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

// retryablePGCodes contains PostgreSQL SQLSTATE codes that are considered transient
// and safe to retry during transaction begin/execute flow.
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

// dbRetryDelays defines backoff delays between retry attempts in WithTxRetry.
var dbRetryDelays = []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second}

// Tx is an alias for pgx.Tx used in repository helpers and signatures.
type Tx = pgx.Tx

// WithTxRetry executes fn in a database transaction with retry logic for transient
// PostgreSQL/transport errors.
//
// Behavior:
// - starts a transaction via pool.Begin
// - executes fn(tx)
// - commits on success
// - rolls back on fn error
// - retries begin/fn failures only when isRetryablePGTransportError returns true
// - stops immediately on non-retryable errors or context cancellation/deadline
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

// isRetryablePGTransportError reports whether err represents a transient PostgreSQL
// or transport-level failure that should be retried.
// It returns false for context cancellation/deadline errors.
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

// sleep waits for duration d or returns context error if ctx is done earlier.
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
