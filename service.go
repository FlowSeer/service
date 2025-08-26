package service

import "context"

// Service represents a long-running component with a well-defined lifecycle.
// It provides methods for initialization, execution, and graceful shutdown,
// as well as accessors for metadata and current state.
type Service interface {
	// Name returns the unique name of the service.
	// This should be a short, human-readable identifier.
	Name() string
	// Version returns the version of the service implementation.
	// This should be a semantic version (e.g., "1.2.3") or a date-based version (e.g., "2024-06-01").
	Version() string
	// Stage returns the current lifecycle stage of the service.
	// This can be used to monitor or orchestrate service state transitions.
	Stage() Stage
	// Initialize prepares the service for execution.
	// This method should be called before Run, and may perform setup such as resource allocation,
	// configuration loading, or dependency checks.
	// The provided context can be used to cancel initialization early.
	// Returns an error if initialization fails or is cancelled.
	Initialize(context.Context) error
	// Run starts the main execution loop of the service.
	// This method should block until the service is stopped, fails, or the context is cancelled.
	// If the context is cancelled, the service must begin a graceful shutdown.
	// Returns an error if the service fails or is interrupted.
	Run(Handle) error
	// Shutdown gracefully stops the service and releases all resources.
	// The provided context can be used to enforce a shutdown deadline or cancellation.
	// After successful shutdown, the service must transition to the StageFinished state.
	// If the context is cancelled or times out before shutdown completes, context.Canceled should be returned.
	Shutdown(ctx context.Context) error
}
