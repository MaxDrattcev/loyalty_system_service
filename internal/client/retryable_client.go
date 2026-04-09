package client

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-resty/resty/v2"
	"net"
	"net/http"
	"os"
	"strconv"
	"syscall"
	"time"
)

var retryDelays = []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second}

type RetryableClient struct {
	Client *resty.Client
}

func NewRetryableClient() *RetryableClient {
	client := resty.New()
	return &RetryableClient{
		Client: client,
	}
}

func (r *RetryableClient) GetWithRetry(ctx context.Context, url string, headers map[string]string) (*resty.Response, error) {
	var lastResp *resty.Response
	var lastErr error

	for attempt := 0; ; attempt++ {
		req := r.Client.R().SetContext(ctx)
		for k, v := range headers {
			req.SetHeader(k, v)
		}
		resp, err := req.Get(url)
		if resp != nil && resp.StatusCode() == http.StatusTooManyRequests {
			if attempt >= len(retryDelays) {
				return resp, fmt.Errorf("too many retries, last status: %d", resp.StatusCode())
			}
			ra := resp.Header().Get("Retry-After")
			raInt, err := strconv.ParseInt(ra, 10, 64)
			if err != nil || raInt <= 0 {
				raInt = 60
			}
			t := time.NewTimer(time.Duration(raInt) * time.Second)
			select {
			case <-ctx.Done():
				t.Stop()
				return nil, ctx.Err()
			case <-t.C:
				t.Stop()
				continue
			}
		}
		lastResp, lastErr = resp, err

		needRetry := false

		if err != nil {
			needRetry = isRetryableTransportError(err)
		} else if resp != nil {
			sc := resp.StatusCode()
			if sc == 204 || sc >= 500 {
				needRetry = true
			}
		}
		if !needRetry {
			return lastResp, lastErr
		}
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		if attempt >= len(retryDelays) {
			return lastResp, lastErr
		}
		t := time.NewTimer(retryDelays[attempt])
		select {
		case <-ctx.Done():
			t.Stop()
			return nil, ctx.Err()
		case <-t.C:
			t.Stop()
		}
	}
}

func isRetryableTransportError(err error) bool {
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}

	var ne net.Error
	if errors.As(err, &ne) {
		if ne.Timeout() {
			return true
		}
	}
	var oe *net.OpError
	if errors.As(err, &oe) {
		return true
	}

	var ose *os.SyscallError
	if errors.As(err, &ose) {
		return isRetryableTransportError(ose.Err)
	}

	var se syscall.Errno
	if errors.As(err, &se) {
		switch se {
		case syscall.ECONNREFUSED, syscall.ECONNRESET, syscall.EPIPE, syscall.ETIMEDOUT, syscall.EHOSTUNREACH, syscall.ENETUNREACH:
			return true
		}
	}
	return false
}
