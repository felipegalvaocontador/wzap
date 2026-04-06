package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	CWMessagesSent = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "cw_messages_sent_total",
		Help: "Total Chatwoot messages sent successfully",
	}, []string{"session", "type", "direction"})

	CWMessagesFailed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "cw_messages_failed_total",
		Help: "Total Chatwoot messages failed",
	}, []string{"session", "type", "error"})

	CWMessageLatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "cw_message_latency_seconds",
		Help:    "Chatwoot message processing latency",
		Buckets: prometheus.DefBuckets,
	}, []string{"session", "type", "direction"})

	CWRetryCount = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "cw_retry_count",
		Help:    "Number of retries per message",
		Buckets: []float64{0, 1, 2, 5, 10, 20, 50},
	}, []string{"session"})

	CWCircuitBreakerState = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cw_circuit_breaker_state",
		Help: "Chatwoot circuit breaker state per session (0=closed, 1=open, 2=half-open)",
	}, []string{"session"})

	CWMediaDownloadBytes = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "cw_media_download_bytes",
		Help: "Total bytes downloaded/uploaded for Chatwoot media",
	}, []string{"session", "type"})

	CWQueueDepth = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cw_queue_depth",
		Help: "Current depth of Chatwoot NATS streams",
	}, []string{"stream"})

	CWDeadLetterCount = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "cw_dead_letter_count",
		Help: "Total messages sent to Chatwoot dead letter queue",
	}, []string{"stream"})

	CWIdempotentDrops = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "cw_idempotent_drops_total",
		Help: "Total duplicate messages dropped by idempotency check",
	}, []string{"session"})

	CWHistoryImportProgress = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cw_history_import_progress",
		Help: "History import progress per session (0-100)",
	}, []string{"session"})
)
