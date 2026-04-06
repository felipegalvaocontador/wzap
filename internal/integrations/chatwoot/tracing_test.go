package chatwoot

import (
	"context"
	"os"
	"testing"

	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func setupTestTracer() (*tracetest.SpanRecorder, func()) {
	sr := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(sr))
	otel.SetTracerProvider(tp)
	return sr, func() { _ = tp.Shutdown(context.Background()) }
}

func TestStartSpan_CreatesSpan(t *testing.T) {
	sr, cleanup := setupTestTracer()
	defer cleanup()

	ctx, span := startSpan(context.Background(), "test.operation",
		spanAttrs("sess-1", "text", "inbound")...)
	_ = ctx
	span.End()

	spans := sr.Ended()
	if len(spans) == 0 {
		t.Fatal("expected at least one span recorded")
	}
	if spans[0].Name() != "test.operation" {
		t.Errorf("expected span name 'test.operation', got %s", spans[0].Name())
	}
}

func TestSpanAttrs_ContainsExpectedFields(t *testing.T) {
	attrs := spanAttrs("my-session", "image", "inbound")
	attrMap := make(map[string]string)
	for _, a := range attrs {
		attrMap[string(a.Key)] = a.Value.AsString()
	}

	checks := map[string]string{
		"messaging.system":  "whatsapp",
		"session.id":        "my-session",
		"message.type":      "image",
		"message.direction": "inbound",
	}
	for k, v := range checks {
		if attrMap[k] != v {
			t.Errorf("expected attr %s=%s, got %s", k, v, attrMap[k])
		}
	}
}

func TestInitTracing_DisabledViaEnv(t *testing.T) {
	t.Setenv("OTEL_SDK_DISABLED", "true")
	defer func() { _ = os.Unsetenv("OTEL_SDK_DISABLED") }()

	shutdown, err := InitTracing(context.Background())
	if err != nil {
		t.Fatalf("expected no error with OTEL_SDK_DISABLED=true, got %v", err)
	}
	if err := shutdown(context.Background()); err != nil {
		t.Errorf("unexpected shutdown error: %v", err)
	}
}

func TestNATSHeaderCarrier_SetGet(t *testing.T) {
	import_nats_pkg := func() {
		// Carrier implements Get/Set/Keys
		headers := make(natsHeaderCarrier)
		headers.Set("traceparent", "00-abc-def-01")
		val := headers.Get("traceparent")
		if val != "00-abc-def-01" {
			t.Errorf("expected '00-abc-def-01', got %s", val)
		}
		keys := headers.Keys()
		if len(keys) != 1 {
			t.Errorf("expected 1 key, got %d", len(keys))
		}
	}
	import_nats_pkg()
}
