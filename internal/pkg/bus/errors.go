package bus

import (
	"fmt"
	"time"
)

type RetryAfterError struct {
	duration time.Duration
}

func NewRetryAfterError(duration time.Duration) *RetryAfterError {
	return &RetryAfterError{duration: duration}
}

func (e *RetryAfterError) Error() string {
	return fmt.Sprintf("retry after %s", e.duration)
}

func (e *RetryAfterError) Duration() time.Duration {
	return e.duration
}
