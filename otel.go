package service

import (
	"context"
	"os"
	"time"

	"github.com/FlowSeer/fail"
	"go.opentelemetry.io/contrib/exporters/autoexport"
	"go.opentelemetry.io/otel/metric"
	metricNoop "go.opentelemetry.io/otel/metric/noop"
	metricSdk "go.opentelemetry.io/otel/sdk/metric"
	traceSdk "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	traceNoop "go.opentelemetry.io/otel/trace/noop"
)

const (
	OtelEnableEnvVar       = "OTEL_ENABLED"
	InstrumentationName    = "github.com/FlowSeer/service"
	InstrumentationVersion = "0.0.1"
)

// tracerKey is the context key type for storing the default service Tracer in a context.
type tracerKey struct{}

// meterKey is the context key type for storing the default service Meter in a context.
type meterKey struct{}

// tracerProviderKey is the context key type for storing a TracerProvider in a context.
type tracerProviderKey struct{}

// meterProviderKey is the context key type for storing a MeterProvider in a context.
type meterProviderKey struct{}

// OtelShutdownFunc is a function that shuts down an OpenTelemetry component.
type OtelShutdownFunc func(context.Context) error

// WithTracer returns a new context with the specified OpenTelemetry Tracer attached.
// This allows downstream code to retrieve the Tracer using Tracer(ctx).
func WithTracer(ctx context.Context, tracer trace.Tracer) context.Context {
	return context.WithValue(ctx, tracerKey{}, tracer)
}

// WithTracerProvider returns a new context with the specified OpenTelemetry TracerProvider attached.
// This allows downstream code to retrieve the TracerProvider using TracerProvider(ctx).
func WithTracerProvider(ctx context.Context, provider trace.TracerProvider) context.Context {
	return context.WithValue(ctx, tracerProviderKey{}, provider)
}

func Tracer(ctx context.Context) trace.Tracer {
	if tracer, ok := ctx.Value(tracerKey{}).(trace.Tracer); ok {
		return tracer
	}

	return traceNoop.NewTracerProvider().Tracer(InstrumentationName, trace.WithInstrumentationVersion(InstrumentationVersion))
}

// TracerProvider retrieves the OpenTelemetry TracerProvider from the context, if present.
// If no TracerProvider is set in the context, a no-op TracerProvider is returned.
func TracerProvider(ctx context.Context) trace.TracerProvider {
	if provider, ok := ctx.Value(tracerProviderKey{}).(trace.TracerProvider); ok {
		return provider
	}
	return traceNoop.NewTracerProvider()
}

// TracerProviderFromEnv constructs a new OpenTelemetry TracerProvider using the provided options.
// This is a convenience for initializing a TracerProvider, e.g., from environment configuration.
func TracerProviderFromEnv(ctx context.Context, opts ...traceSdk.TracerProviderOption) (trace.TracerProvider, OtelShutdownFunc, error) {
	exporter, err := autoexport.NewSpanExporter(ctx)
	if err != nil {
		return traceNoop.NewTracerProvider(), OtelNoopShutdown, fail.New().
			Context(ctx).
			Cause(err).
			Msg("failed to create OTEL span exporter")
	}

	return traceSdk.NewTracerProvider(append(opts,
		traceSdk.WithBatcher(exporter,
			traceSdk.WithBatchTimeout(1*time.Second),
		),
	)...), exporter.Shutdown, nil
}

// WithMeter returns a new context with the specified OpenTelemetry Meter attached.
// This allows downstream code to retrieve the Meter using Meter(ctx).
func WithMeter(ctx context.Context, meter metric.Meter) context.Context {
	return context.WithValue(ctx, meterKey{}, meter)
}

// WithMeterProvider returns a new context with the specified OpenTelemetry MeterProvider attached.
// This allows downstream code to retrieve the MeterProvider using MeterProvider(ctx).
func WithMeterProvider(ctx context.Context, provider metric.MeterProvider) context.Context {
	return context.WithValue(ctx, meterProviderKey{}, provider)
}

// Meter retrieves the OpenTelemetry Meter from the context, if present.
// If no Meter is set in the context, a Meter is created using MeterProvider(ctx).
func Meter(ctx context.Context) metric.Meter {
	if meter, ok := ctx.Value(meterKey{}).(metric.Meter); ok {
		return meter
	}

	return MeterProvider(ctx).Meter(InstrumentationName, metric.WithInstrumentationVersion(InstrumentationVersion))
}

// MeterProvider retrieves the OpenTelemetry MeterProvider from the context, if present.
// If no MeterProvider is set in the context, a no-op MeterProvider is returned.
func MeterProvider(ctx context.Context) metric.MeterProvider {
	if provider, ok := ctx.Value(meterProviderKey{}).(metric.MeterProvider); ok {
		return provider
	}
	return metricNoop.NewMeterProvider()
}

// MeterProviderFromEnv constructs a new OpenTelemetry MeterProvider using the provided options.
// This is a convenience for initializing a MeterProvider, e.g., from environment configuration.
func MeterProviderFromEnv(ctx context.Context, opts ...metricSdk.Option) (metric.MeterProvider, OtelShutdownFunc, error) {
	reader, err := autoexport.NewMetricReader(ctx)
	if err != nil {
		return metricNoop.NewMeterProvider(), OtelNoopShutdown, fail.New().
			Context(ctx).
			Cause(err).
			Msg("failed to create OTEL metric exporter")
	}

	return metricSdk.NewMeterProvider(append(opts,
		metricSdk.WithReader(reader),
	)...), reader.Shutdown, nil
}

// IsOtelEnabled checks whether OpenTelemetry instrumentation is enabled by looking for an
// environment variable named {PREFIX}_OTEL_ENABLED (normalized using EnvName).
// Returns true if the variable is set, false otherwise.
func IsOtelEnabled(prefix string) bool {
	_, ok := os.LookupEnv(EnvName(prefix, OtelEnableEnvVar))
	return ok
}

// OtelNoopShutdown is a no-op OtelShutdownFunc.
func OtelNoopShutdown(context.Context) error {
	return nil
}
