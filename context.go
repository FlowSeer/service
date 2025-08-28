package service

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

// Context is a wrapper around context.Context that provides
// convenient access to logging, tracing, and metrics providers.
// It is intended to be used within service implementations to
// access observability primitives in a unified way.
type Context struct {
	context.Context

	logger *slog.Logger

	tracerProvider trace.TracerProvider
	defaultTracer  trace.Tracer

	meterProvider metric.MeterProvider
	defaultMeter  metric.Meter
}

// Logger returns the slog.Logger associated with this Context.
func (c *Context) Logger() *slog.Logger {
	return c.logger
}

// TracerProvider returns the OpenTelemetry TracerProvider associated with this Context.
func (c *Context) TracerProvider() trace.TracerProvider {
	return c.tracerProvider
}

// Tracer returns the default OpenTelemetry Tracer for this Context.
func (c *Context) Tracer() trace.Tracer {
	return c.defaultTracer
}

// MeterProvider returns the OpenTelemetry MeterProvider associated with this Context.
func (c *Context) MeterProvider() metric.MeterProvider {
	return c.meterProvider
}

// Meter returns the default OpenTelemetry Meter for this Context.
func (c *Context) Meter() metric.Meter {
	return c.defaultMeter
}

// Info logs an informational message using the Context's logger.
func (c *Context) Info(msg string, args ...interface{}) {
	c.logger.Info(msg, args...)
}

// Warn logs a warning message using the Context's logger.
func (c *Context) Warn(msg string, args ...interface{}) {
	c.logger.Warn(msg, args...)
}

// Error logs an error message using the Context's logger.
func (c *Context) Error(msg string, args ...interface{}) {
	c.logger.Error(msg, args...)
}
