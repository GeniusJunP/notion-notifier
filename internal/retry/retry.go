package retry

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"
)

type Config struct {
	MaxRetries int
	BaseDelay  time.Duration
	MaxDelay   time.Duration
}

func (c Config) WithDefaults() Config {
	if c.MaxRetries <= 0 {
		c.MaxRetries = 3
	}
	if c.BaseDelay <= 0 {
		c.BaseDelay = time.Second
	}
	if c.MaxDelay <= 0 {
		c.MaxDelay = 30 * time.Second
	}
	return c
}

func IsRetryableStatus(status int) bool {
	return status == http.StatusTooManyRequests || status >= 500
}

func BackoffDelay(cfg Config, attempt int, retryAfter time.Duration) time.Duration {
	cfg = cfg.WithDefaults()
	if retryAfter > 0 {
		if retryAfter > cfg.MaxDelay {
			return cfg.MaxDelay
		}
		return retryAfter
	}
	delay := cfg.BaseDelay * time.Duration(1<<attempt)
	if delay > cfg.MaxDelay {
		return cfg.MaxDelay
	}
	return delay
}

func Sleep(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return nil
	}
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func ParseRetryAfter(header string) (time.Duration, bool) {
	if header == "" {
		return 0, false
	}
	if seconds, err := strconv.Atoi(header); err == nil {
		return time.Duration(seconds) * time.Second, true
	}
	if t, err := http.ParseTime(header); err == nil {
		return time.Until(t), true
	}
	return 0, false
}

var ErrRetriesExhausted = errors.New("retries exhausted")
