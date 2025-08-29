package service

import (
	"context"
	"log/slog"
	"os"
	"strings"
)

//go:generate go tool golang.org/x/tools/cmd/stringer -type LogFormat -trimprefix LogFormat

// LogFormat specifies the output format for logs.
type LogFormat int

const (
	// LogFormatJson outputs logs in JSON format.
	LogFormatJson LogFormat = iota
	// LogFormatText outputs logs in human-readable text format.
	LogFormatText
)

// Context key types for storing log configuration in context.
type logLevelKey struct{}
type logFormatKey struct{}
type loggerKey struct{}

// WithLogLevel returns a new context with the specified log level.
// The level parameter overrides the default log level for all operations
// performed within the returned context.
func WithLogLevel(ctx context.Context, level slog.Leveler) context.Context {
	return context.WithValue(ctx, logLevelKey{}, level)
}

// LogLevel retrieves the log level from the context.
// If no log level is set in the context, slog.LevelInfo is returned as the default.
func LogLevel(ctx context.Context) slog.Leveler {
	if level, ok := ctx.Value(logLevelKey{}).(slog.Leveler); ok {
		return level
	}
	return slog.LevelInfo
}

// LogLevelFromEnv reads the log level from environment variables.
// The prefix parameter is used to namespace the environment variable.
// If prefix is provided, it will look for {PREFIX}_LOG_LEVEL.
// If prefix is empty, it will look for SERVICE_LOG_LEVEL.
// Valid values are: "debug", "info", "warn", "error" (case-insensitive).
// Returns slog.LevelInfo as the default if no valid level is found.
func LogLevelFromEnv(prefix string) slog.Leveler {
	switch strings.ToLower(os.Getenv(EnvName(prefix, "LOG_LEVEL"))) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	}

	return slog.LevelInfo
}

// WithLogFormat returns a new context with the specified log format.
// The format parameter determines how log messages will be formatted
// for all operations performed within the returned context.
func WithLogFormat(ctx context.Context, format LogFormat) context.Context {
	return context.WithValue(ctx, logFormatKey{}, format)
}

// LogFormatFromEnv returns the log format to use, as determined by an environment variable.
// The 'prefix' parameter is used to namespace the environment variable name. If 'prefix' is non-empty,
// the function looks for an environment variable named {PREFIX}_LOG_FORMAT (ensuring a trailing underscore if needed).
// If 'prefix' is empty, it defaults to SERVICE_LOG_FORMAT.
//
// Recognized values for the environment variable (case-insensitive) are:
//   - "text", "pretty", "console": all map to LogFormatText
//   - "json", "structured": both map to LogFormatJson
//
// If the environment variable is unset or contains an unrecognized value, LogFormatJson is returned as the default.
func LogFormatFromEnv(prefix string) LogFormat {
	if prefix == "" {
		prefix = "SERVICE_"
	} else {
		if !strings.HasSuffix(prefix, "_") {
			prefix += "_"
		}
	}

	switch strings.ToLower(os.Getenv(EnvName(prefix, "LOG_FORMAT"))) {
	case "text", "pretty", "console":
		return LogFormatText
	case "json", "structured":
		return LogFormatJson
	}

	return LogFormatJson
}

func WithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey{}, logger)
}

func LoggerFromEnv(prefix string) *slog.Logger {
	level := LogLevelFromEnv(prefix)
	format := LogFormatFromEnv(prefix)

	var handler slog.Handler
	switch format {
	case LogFormatText:
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: level,
		})
	default:
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: level,
		})
	}

	return slog.New(handler)
}
