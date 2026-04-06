package chatwoot

import (
	"testing"

	"wzap/internal/metrics"
)

func TestMetrics_PrometheusRegistered(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("metrics operation panicked: %v", r)
		}
	}()

	metrics.CWMessagesSent.WithLabelValues("test-sess", "text", "inbound").Inc()
	metrics.CWMessagesFailed.WithLabelValues("test-sess", "text", "error").Inc()
	metrics.CWMessageLatency.WithLabelValues("test-sess", "text", "inbound").Observe(0.1)
	metrics.CWRetryCount.WithLabelValues("test-sess").Observe(2)
	metrics.CWMediaDownloadBytes.WithLabelValues("test-sess", "image").Add(1024)
	metrics.CWQueueDepth.WithLabelValues("CW_INBOUND").Set(5)
	metrics.CWQueueDepth.WithLabelValues("CW_OUTBOUND").Set(3)
	metrics.CWDeadLetterCount.WithLabelValues("inbound").Inc()
	metrics.CWIdempotentDrops.WithLabelValues("test-sess").Inc()
	metrics.CWHistoryImportProgress.WithLabelValues("test-sess").Set(50)
	metrics.CWCircuitBreakerState.WithLabelValues("test-sess").Set(0)
}

func TestMetrics_CircuitBreakerStateTransitions(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("panic: %v", r)
		}
	}()

	metrics.CWCircuitBreakerState.WithLabelValues("cb-test").Set(float64(cbClosed))
	metrics.CWCircuitBreakerState.WithLabelValues("cb-test").Set(float64(cbOpen))
	metrics.CWCircuitBreakerState.WithLabelValues("cb-test").Set(float64(cbHalfOpen))
	metrics.CWCircuitBreakerState.WithLabelValues("cb-test").Set(float64(cbClosed))
}
