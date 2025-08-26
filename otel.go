package service

import (
	"context"
	"os"

	"go.opentelemetry.io/otel/metric"
	metricNoop "go.opentelemetry.io/otel/metric/noop"
	metricSdk "go.opentelemetry.io/otel/sdk/metric"
	traceSdk "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	traceNoop "go.opentelemetry.io/otel/trace/noop"
)

// tracerProviderKey is the context key type for storing a TracerProvider in a context.
type tracerProviderKey struct{}

// meterProviderKey is the context key type for storing a MeterProvider in a context.
type meterProviderKey struct{}

// WithTracerProvider returns a new context with the specified OpenTelemetry TracerProvider attached.
// This allows downstream code to retrieve the TracerProvider using TracerProvider(ctx).
func WithTracerProvider(ctx context.Context, provider trace.TracerProvider) context.Context {
	return context.WithValue(ctx, tracerProviderKey{}, provider)
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
func TracerProviderFromEnv(opts ...traceSdk.TracerProviderOption) trace.TracerProvider {
	return traceSdk.NewTracerProvider(opts...)
}

// WithMeterProvider returns a new context with the specified OpenTelemetry MeterProvider attached.
// This allows downstream code to retrieve the MeterProvider using MeterProvider(ctx).
func WithMeterProvider(ctx context.Context, provider metric.MeterProvider) context.Context {
	return context.WithValue(ctx, meterProviderKey{}, provider)
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
func MeterProviderFromEnv(opts ...metricSdk.Option) metric.MeterProvider {
	return metricSdk.NewMeterProvider(opts...)
}

// IsOtelEnabled checks whether OpenTelemetry instrumentation is enabled by looking for an
// environment variable named {PREFIX}_OTEL_ENABLED (normalized using envName).
// Returns true if the variable is set, false otherwise.
func IsOtelEnabled(prefix string) bool {
	_, ok := os.LookupEnv(envName(prefix, "OTEL_ENABLED"))
	return ok
}
