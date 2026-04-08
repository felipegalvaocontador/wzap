package chatwoot

import (
	"context"
	"errors"
	"strings"
)

func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return true
	}

	msg := err.Error()

	if strings.Contains(msg, "status=429") {
		return true
	}

	if strings.Contains(msg, "chatwoot API error:") {
		return strings.Contains(msg, "status=5")
	}

	if strings.Contains(msg, "do request:") {
		return true
	}

	return false
}
