package retry

import (
	"context"
	"fmt"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
	"go.uber.org/zap"
)

// defaultRetryableErrors lists Telegram errors that are safe to retry.
// These are transient server-side or network-level failures.
var defaultRetryableErrors = []string{
	"Timedout",
	"No workers running",
	"RPC_CALL_FAIL",
	"RPC_MCGET_FAIL",
	"WORKER_BUSY_TOO_LONG_RETRY",
	"memory limit exit",
}

type retryMiddleware struct {
	max    int
	errors []string
	logger *zap.Logger
}

// Handle implements telegram.Middleware.
func (r retryMiddleware) Handle(next tg.Invoker) telegram.InvokeFunc {
	return func(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
		retries := 0

		for retries < r.max {
			if err := next.Invoke(ctx, input, output); err != nil {
				if tgerr.Is(err, r.errors...) {
					r.logger.Debug("download retry",
						zap.Int("retries", retries),
						zap.Error(err),
					)
					retries++
					continue
				}
				return fmt.Errorf("retry middleware skip: %w", err)
			}
			return nil
		}

		return fmt.Errorf("retry limit reached after %d attempts", r.max)
	}
}

// New returns a middleware that retries requests on transient Telegram errors.
// Default retryable errors are included; extraErrors are appended for customization.
func New(max int, logger *zap.Logger, extraErrors ...string) telegram.Middleware {
	return retryMiddleware{
		max:    max,
		errors: append(defaultRetryableErrors, extraErrors...),
		logger: logger,
	}
}
