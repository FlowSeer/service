package service

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

// Handle provides access to contextual resources for a running service.
// It encapsulates the context, logger, and OpenTelemetry providers for metrics and tracing.
type Handle struct {
	ctx            context.Context
	logger         *slog.Logger
	meterProvider  metric.MeterProvider
	tracerProvider trace.TracerProvider
}

// Context returns the base context associated with the service.
// This context is typically used for cancellation and propagation of deadlines.
func (c *Handle) Context() context.Context {
	return c.ctx
}

// MeterProvider returns the OpenTelemetry MeterProvider used for metrics instrumentation.
func (c *Handle) MeterProvider() metric.MeterProvider {
	return c.meterProvider
}

// Meter returns a named OpenTelemetry Meter for recording metrics.
// The name should identify the instrumentation scope.
func (c *Handle) Meter(name string) metric.Meter {
	return c.meterProvider.Meter(name)
}

// TracerProvider returns the OpenTelemetry TracerProvider used for distributed tracing.
func (c *Handle) TracerProvider() trace.TracerProvider {
	return c.tracerProvider
}

// Tracer returns a named OpenTelemetry Tracer for creating spans.
// The name should identify the instrumentation scope.
func (c *Handle) Tracer(name string) trace.Tracer {
	return c.tracerProvider.Tracer(name)
}

// Logger returns the logger associated with the service.
// This logger should be used for all structured logging within the service.
func (c *Handle) Logger() *slog.Logger {
	return c.logger
}
