package utils

import (
	"context"
	"time"
)

// RetryOptions 定义重试操作的选项
type RetryOptions struct {
	MaxRetries  int
	BaseDelay   time.Duration
	MaxDelay    time.Duration
	MaxDuration time.Duration
}

// DefaultRetryOptions 返回默认的重试选项
func DefaultRetryOptions() RetryOptions {
	return RetryOptions{
		MaxRetries:  3,
		BaseDelay:   100 * time.Millisecond,
		MaxDelay:    1 * time.Second,
		MaxDuration: 5 * time.Second,
	}
}

// RetryWithContext 执行带有重试机制的操作
func RetryWithContext(ctx context.Context, opts RetryOptions, operation func() error) error {
	var lastErr error
	delay := opts.BaseDelay

	for retries := 0; retries < opts.MaxRetries; retries++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := operation(); err == nil {
				return nil
			} else {
				lastErr = err
			}
		}

		if retries < opts.MaxRetries-1 {
			time.Sleep(delay)
			delay *= 2
			if delay > opts.MaxDelay {
				delay = opts.MaxDelay
			}
		}
	}

	return lastErr
}