package repository

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgconn"
	"github.com/stretchr/testify/require"
)

func TestIsRetryablePGTransportError_TableDriven(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "context canceled is not retryable",
			err:  context.Canceled,
			want: false,
		},
		{
			name: "context deadline exceeded is not retryable",
			err:  context.DeadlineExceeded,
			want: false,
		},
		{
			name: "pg error 08xxx is retryable",
			err:  &pgconn.PgError{Code: "08006"},
			want: true,
		},
		{
			name: "pg error from retryable list 40001 is retryable",
			err:  &pgconn.PgError{Code: "40001"},
			want: true,
		},
		{
			name: "pg error from retryable list 57P03 is retryable",
			err:  &pgconn.PgError{Code: "57P03"},
			want: true,
		},
		{
			name: "wrapped retryable pg error is retryable",
			err:  fmt.Errorf("wrapped: %w", &pgconn.PgError{Code: "40P01"}),
			want: true,
		},
		{
			name: "pg error not in retryable set is not retryable",
			err:  &pgconn.PgError{Code: "23505"},
			want: false,
		},
		{
			name: "generic error is not retryable",
			err:  errors.New("some error"),
			want: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := isRetryablePGTransportError(tt.err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestSleep_TableDriven(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		cancelCtx bool
		wantErr   error
	}{
		{
			name:      "returns nil when timer fires",
			cancelCtx: false,
			wantErr:   nil,
		},
		{
			name:      "returns context canceled when context canceled",
			cancelCtx: true,
			wantErr:   context.Canceled,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var (
				ctx    context.Context
				cancel context.CancelFunc
			)

			if tt.cancelCtx {
				ctx, cancel = context.WithCancel(context.Background())
				cancel()
			} else {
				ctx, cancel = context.WithTimeout(context.Background(), 100*time.Millisecond)
			}
			defer cancel()

			err := sleep(ctx, 5*time.Millisecond)

			if tt.wantErr == nil {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, tt.wantErr)
			}
		})
	}
}
