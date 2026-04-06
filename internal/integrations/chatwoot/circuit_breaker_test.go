package chatwoot

import (
	"testing"
	"time"
)

func TestCircuitBreaker_ClosedByDefault(t *testing.T) {
	cb := newCircuitBreaker()
	if !cb.Allow() {
		t.Error("expected circuit breaker to allow requests in CLOSED state")
	}
	if cb.State() != cbClosed {
		t.Errorf("expected state CLOSED, got %d", cb.State())
	}
}

func TestCircuitBreaker_OpenAfterThreshold(t *testing.T) {
	cb := newCircuitBreaker()
	for i := 0; i < cbThreshold; i++ {
		cb.RecordFailure()
	}
	if cb.State() != cbOpen {
		t.Errorf("expected state OPEN after %d failures, got %d", cbThreshold, cb.State())
	}
	if cb.Allow() {
		t.Error("expected circuit breaker to reject requests in OPEN state")
	}
}

func TestCircuitBreaker_HalfOpenAfterTimeout(t *testing.T) {
	cb := newCircuitBreaker()
	for i := 0; i < cbThreshold; i++ {
		cb.RecordFailure()
	}
	cb.mu.Lock()
	cb.lastFail = time.Now().Add(-(cbTimeout + time.Second))
	cb.mu.Unlock()

	if !cb.Allow() {
		t.Error("expected circuit breaker to allow in HALF_OPEN after timeout")
	}
	if cb.State() != cbHalfOpen {
		t.Errorf("expected state HALF_OPEN, got %d", cb.State())
	}
}

func TestCircuitBreaker_ClosedAfterSuccessInHalfOpen(t *testing.T) {
	cb := newCircuitBreaker()
	for i := 0; i < cbThreshold; i++ {
		cb.RecordFailure()
	}
	cb.mu.Lock()
	cb.lastFail = time.Now().Add(-(cbTimeout + time.Second))
	cb.mu.Unlock()
	cb.Allow() // transitions to HALF_OPEN

	cb.RecordSuccess()
	if cb.State() != cbClosed {
		t.Errorf("expected state CLOSED after success in HALF_OPEN, got %d", cb.State())
	}
	if !cb.Allow() {
		t.Error("expected circuit breaker to allow requests after recovery")
	}
}

func TestCircuitBreaker_BackToOpenOnFailureInHalfOpen(t *testing.T) {
	cb := newCircuitBreaker()
	for i := 0; i < cbThreshold; i++ {
		cb.RecordFailure()
	}
	cb.mu.Lock()
	cb.lastFail = time.Now().Add(-(cbTimeout + time.Second))
	cb.mu.Unlock()
	cb.Allow() // transitions to HALF_OPEN

	cb.RecordFailure()
	if cb.State() != cbOpen {
		t.Errorf("expected state OPEN after failure in HALF_OPEN, got %d", cb.State())
	}
}

func TestCircuitBreakerManager_PerSession(t *testing.T) {
	mgr := newCircuitBreakerManager()

	if !mgr.Allow("session-1") {
		t.Error("expected session-1 to be allowed initially")
	}
	if !mgr.Allow("session-2") {
		t.Error("expected session-2 to be allowed initially")
	}

	for i := 0; i < cbThreshold; i++ {
		mgr.RecordFailure("session-1")
	}

	if mgr.Allow("session-1") {
		t.Error("expected session-1 to be blocked after threshold failures")
	}
	if !mgr.Allow("session-2") {
		t.Error("expected session-2 to still be allowed (independent circuit breaker)")
	}
}

func TestCircuitBreakerManager_RecordSuccess(t *testing.T) {
	mgr := newCircuitBreakerManager()
	for i := 0; i < cbThreshold; i++ {
		mgr.RecordFailure("sess")
	}
	mgr.get("sess").mu.Lock()
	mgr.get("sess").lastFail = time.Now().Add(-(cbTimeout + time.Second))
	mgr.get("sess").mu.Unlock()
	mgr.Allow("sess") // HALF_OPEN

	mgr.RecordSuccess("sess")
	if !mgr.Allow("sess") {
		t.Error("expected sess to be allowed after recovery")
	}
}
