package chatwoot

import (
	"context"
	"fmt"
	"testing"
)

func TestIsRetryableError_Nil(t *testing.T) {
	if isRetryableError(nil) {
		t.Error("expected false for nil error")
	}
}

func TestIsRetryableError_Status500(t *testing.T) {
	err := fmt.Errorf("chatwoot API error: status=500, body=internal server error")
	if !isRetryableError(err) {
		t.Error("expected true for status=500")
	}
}

func TestIsRetryableError_Status429(t *testing.T) {
	err := fmt.Errorf("chatwoot API error: status=429, body=too many requests")
	if !isRetryableError(err) {
		t.Error("expected true for status=429")
	}
}

func TestIsRetryableError_ContextTimeout(t *testing.T) {
	if !isRetryableError(context.DeadlineExceeded) {
		t.Error("expected true for context.DeadlineExceeded")
	}
	if !isRetryableError(context.Canceled) {
		t.Error("expected true for context.Canceled")
	}
}

func TestIsRetryableError_Status404(t *testing.T) {
	err := fmt.Errorf("chatwoot API error: status=404, body=not found")
	if isRetryableError(err) {
		t.Error("expected false for status=404")
	}
}

func TestIsRetryableError_Status422(t *testing.T) {
	err := fmt.Errorf("chatwoot API error: status=422, body=unprocessable entity")
	if isRetryableError(err) {
		t.Error("expected false for status=422")
	}
}

func TestIsRetryableError_NetworkError(t *testing.T) {
	err := fmt.Errorf("do request: dial tcp: connection refused")
	if !isRetryableError(err) {
		t.Error("expected true for network/do request error")
	}
}
