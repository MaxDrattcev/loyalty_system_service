package client

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func patchRetryDelays(t *testing.T, d []time.Duration) {
	t.Helper()
	old := retryDelays
	retryDelays = d
	t.Cleanup(func() { retryDelays = old })
}

func TestRetryableClient_GetWithRetry_Table(t *testing.T) {
	tcs := []struct {
		name            string
		retryDelay      []time.Duration
		serverHandler   func(calls *atomic.Int32) http.HandlerFunc
		ctxFactory      func() (context.Context, context.CancelFunc)
		wantErr         bool
		wantErrContains string
		wantNilResp     bool
		wantStatus      int
		wantCalls       int32
	}{
		{
			name:       "200 success",
			retryDelay: []time.Duration{time.Millisecond},
			serverHandler: func(calls *atomic.Int32) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					calls.Add(1)
					w.WriteHeader(http.StatusOK)
				}
			},
			ctxFactory:  func() (context.Context, context.CancelFunc) { return context.Background(), func() {} },
			wantErr:     false,
			wantNilResp: false,
			wantStatus:  http.StatusOK,
			wantCalls:   1,
		},
		{
			name:       "500 then 200",
			retryDelay: []time.Duration{time.Millisecond, time.Millisecond},
			serverHandler: func(calls *atomic.Int32) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					if calls.Add(1) == 1 {
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
					w.WriteHeader(http.StatusOK)
				}
			},
			ctxFactory:  func() (context.Context, context.CancelFunc) { return context.Background(), func() {} },
			wantErr:     false,
			wantNilResp: false,
			wantStatus:  http.StatusOK,
			wantCalls:   2,
		},
		{
			name:       "all 500 returns last response",
			retryDelay: []time.Duration{time.Millisecond, time.Millisecond, time.Millisecond},
			serverHandler: func(calls *atomic.Int32) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					calls.Add(1)
					w.WriteHeader(http.StatusInternalServerError)
				}
			},
			ctxFactory:  func() (context.Context, context.CancelFunc) { return context.Background(), func() {} },
			wantErr:     false,
			wantNilResp: false,
			wantStatus:  http.StatusInternalServerError,
			wantCalls:   4, // attempt 0..3
		},
		{
			name:       "204 retries then returns 204",
			retryDelay: []time.Duration{time.Millisecond, time.Millisecond, time.Millisecond},
			serverHandler: func(calls *atomic.Int32) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					calls.Add(1)
					w.WriteHeader(http.StatusNoContent)
				}
			},
			ctxFactory:  func() (context.Context, context.CancelFunc) { return context.Background(), func() {} },
			wantErr:     false,
			wantNilResp: false,
			wantStatus:  http.StatusNoContent,
			wantCalls:   4,
		},
		{
			name:       "429 too many retries",
			retryDelay: []time.Duration{time.Millisecond}, // len=1 => 2 calls for 429 branch
			serverHandler: func(calls *atomic.Int32) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					calls.Add(1)
					w.Header().Set("Retry-After", "1")
					w.WriteHeader(http.StatusTooManyRequests)
				}
			},
			ctxFactory:      func() (context.Context, context.CancelFunc) { return context.Background(), func() {} },
			wantErr:         true,
			wantErrContains: "too many retries",
			wantNilResp:     false,
			wantStatus:      http.StatusTooManyRequests,
			wantCalls:       2,
		},
		{
			name:       "context canceled while waiting retry-after",
			retryDelay: []time.Duration{time.Millisecond},
			serverHandler: func(calls *atomic.Int32) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					calls.Add(1)
					w.Header().Set("Retry-After", "60")
					w.WriteHeader(http.StatusTooManyRequests)
				}
			},
			ctxFactory: func() (context.Context, context.CancelFunc) {
				ctx, cancel := context.WithCancel(context.Background())
				go func() {
					time.Sleep(5 * time.Millisecond)
					cancel()
				}()
				return ctx, cancel
			},
			wantErr:     true,
			wantNilResp: true,
			wantCalls:   1,
		},
		{
			name:       "context canceled during 500 backoff",
			retryDelay: []time.Duration{50 * time.Millisecond},
			serverHandler: func(calls *atomic.Int32) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					calls.Add(1)
					w.WriteHeader(http.StatusInternalServerError)
				}
			},
			ctxFactory: func() (context.Context, context.CancelFunc) {
				ctx, cancel := context.WithCancel(context.Background())
				go func() {
					time.Sleep(5 * time.Millisecond)
					cancel()
				}()
				return ctx, cancel
			},
			wantErr:     true,
			wantNilResp: true,
			wantCalls:   1,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			patchRetryDelays(t, tc.retryDelay)

			var calls atomic.Int32
			srv := httptest.NewServer(tc.serverHandler(&calls))
			defer srv.Close()

			ctx, cancel := tc.ctxFactory()
			defer cancel()

			c := NewRetryableClient()
			resp, err := c.GetWithRetry(ctx, srv.URL, nil)

			if tc.wantErr {
				require.Error(t, err)
				if tc.wantErrContains != "" {
					require.Contains(t, err.Error(), tc.wantErrContains)
				}
			} else {
				require.NoError(t, err)
			}

			if tc.wantNilResp {
				require.Nil(t, resp)
			} else {
				require.NotNil(t, resp)
				require.Equal(t, tc.wantStatus, resp.StatusCode())
			}

			require.Equal(t, tc.wantCalls, calls.Load())
		})
	}
}

type timeoutErr struct{}

func (timeoutErr) Error() string   { return "timeout" }
func (timeoutErr) Timeout() bool   { return true }
func (timeoutErr) Temporary() bool { return false }

func TestIsRetryableTransportError_Table(t *testing.T) {
	tcs := []struct {
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
			name: "net timeout error is retryable",
			err:  timeoutErr{},
			want: true,
		},
		{
			name: "syscall ECONNREFUSED is retryable",
			err:  syscall.ECONNREFUSED,
			want: true,
		},
		{
			name: "net.OpError is retryable",
			err:  &net.OpError{Op: "dial", Net: "tcp", Err: syscall.ECONNRESET},
			want: true,
		},
		{
			name: "generic error is not retryable",
			err:  errors.New("some error"),
			want: false,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, isRetryableTransportError(tc.err))
		})
	}
}
