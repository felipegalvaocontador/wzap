package chatwoot

import (
	"context"
	"os"
	"strconv"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"

	"github.com/nats-io/nats.go"
)

const tracerName = "wzap/chatwoot"

func InitTracing(ctx context.Context) (func(context.Context) error, error) {
	if os.Getenv("OTEL_SDK_DISABLED") == "true" {
		return func(ctx context.Context) error { return nil }, nil
	}

	exporter, err := otlptracehttp.New(ctx)
	if err != nil {
		return nil, err
	}

	samplerArg := 0.1
	if v := os.Getenv("OTEL_TRACES_SAMPLER_ARG"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			samplerArg = f
		}
	}

	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName("wzap"),
		semconv.ServiceVersion("1.0.0"),
	)

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(samplerArg))),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return tp.Shutdown, nil
}

func startSpan(ctx context.Context, name string, attrs ...attribute.KeyValue) (context.Context, trace.Span) {
	tracer := otel.Tracer(tracerName)
	return tracer.Start(ctx, name, trace.WithAttributes(attrs...))
}

func spanAttrs(sessionID, msgType, direction string) []attribute.KeyValue {
	return []attribute.KeyValue{
		attribute.String("messaging.system", "whatsapp"),
		attribute.String("session.id", sessionID),
		attribute.String("message.type", msgType),
		attribute.String("message.direction", direction),
	}
}

// natsHeaderCarrier adapts nats.Header to the TextMapCarrier interface.
type natsHeaderCarrier nats.Header

func (c natsHeaderCarrier) Get(key string) string {
	return nats.Header(c).Get(key)
}
func (c natsHeaderCarrier) Set(key, val string) {
	nats.Header(c).Set(key, val)
}
func (c natsHeaderCarrier) Keys() []string {
	keys := make([]string, 0, len(c))
	for k := range c {
		keys = append(keys, k)
	}
	return keys
}

// InjectNATSHeaders injects the current trace context into NATS message headers.
func InjectNATSHeaders(ctx context.Context, msg *nats.Msg) {
	if msg.Header == nil {
		msg.Header = make(nats.Header)
	}
	otel.GetTextMapPropagator().Inject(ctx, natsHeaderCarrier(msg.Header))
}

// ExtractNATSContext extracts trace context from NATS message headers.
func ExtractNATSContext(ctx context.Context, msg *nats.Msg) context.Context {
	if msg.Header == nil {
		return ctx
	}
	return otel.GetTextMapPropagator().Extract(ctx, natsHeaderCarrier(msg.Header))
}
