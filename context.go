package service

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel/log"
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
	tracerShutdown OtelShutdownFunc
	defaultTracer  trace.Tracer

	meterProvider metric.MeterProvider
	meterShutdown OtelShutdownFunc

	loggerProvider log.LoggerProvider
	loggerShutdown OtelShutdownFunc

	defaultMeter metric.Meter
}

// LoggerProvider returns the OpenTelemetry LoggerProvider associated with this Context.
//
// This provider is used to create OpenTelemetry-compatible loggers. If OpenTelemetry logging
// is enabled, the slog.Logger returned by Logger() will be configured to bridge logs to this
// LoggerProvider, allowing logs to be exported via OpenTelemetry pipelines.
func (c *Context) LoggerProvider() log.LoggerProvider {
	return c.loggerProvider
}

// Logger returns the slog.Logger instance associated with this Context.
//
// This logger is intended for application logging and may be configured to bridge logs
// to OpenTelemetry if OTEL logging is enabled. Use this logger for all structured logging
// within service implementations to ensure logs are properly captured and exported.
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

// Debug logs an debug message using the Context's logger.
func (c *Context) Debug(msg string, args ...any) {
	c.logger.Debug(msg, args...)
}

// Info logs an informational message using the Context's logger.
func (c *Context) Info(msg string, args ...any) {
	c.logger.Info(msg, args...)
}

// Warn logs a warning message using the Context's logger.
func (c *Context) Warn(msg string, args ...any) {
	c.logger.Warn(msg, args...)
}

// Error logs an error message using the Context's logger.
func (c *Context) Error(msg string, args ...any) {
	c.logger.Error(msg, args...)
}
