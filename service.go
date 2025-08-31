package service

import "sync/atomic"

// Service represents a long-running component with a well-defined lifecycle.
// It provides methods for initialization, execution, and graceful shutdown,
// as well as accessors for metadata and current state.
type Service interface {
	// Name returns the unique name of the service.
	// This should be a short, human-readable identifier.
	Name() string
	// Namespace returns the namespace of the service.
	Namespace() string
	// Version returns the version of the service implementation.
	// This should be a semantic version (e.g., "1.2.3") or a date-based version (e.g., "2024-06-01").
	Version() string
	// Health returns the current health status of the service.
	Health() Health
	// Initialize prepares the service for execution.
	// This method should be called before Run, and may perform setup such as resource allocation,
	// configuration loading, or dependency checks.
	// The provided context can be used to cancel initialization early.
	// Returns an error if initialization fails or is cancelled.
	Initialize(*Context) error
	// Run starts the main execution loop of the service.
	// This method should block until the service is stopped, fails, or the context is cancelled.
	// If the context is cancelled, the service must begin a graceful shutdown.
	// Returns an error if the service fails or is interrupted.
	Run(*Context) error
	// Shutdown gracefully stops the service and releases all resources.
	// The provided context can be used to enforce a shutdown deadline or cancellation.
	// After successful shutdown, the service must transition to the StageFinished state.
	// If the context is cancelled or times out before shutdown completes, context.Canceled should be returned.
	Shutdown(*Context) error
}

// Simple returns a Service implementation with the given name, namespace, version, and run function.
// The returned service is suitable for simple use cases where only a run function is needed.
func Simple(name, namespace, version string, fn func(*Context) error) Service {
	return &simpleService{
		name:      name,
		namespace: namespace,
		version:   version,
		fn:        fn,
	}
}

// simpleService is a basic implementation of the Service interface.
// It is intended for use with the Simple constructor.
type simpleService struct {
	name      string
	namespace string
	version   string
	err       error
	fn        func(*Context) error

	started           atomic.Bool
	stopped           atomic.Bool
	shutdownRequested atomic.Bool
}

// Name returns the unique name of the service.
func (s *simpleService) Name() string {
	return s.name
}

// Namespace returns the namespace of the service.
func (s *simpleService) Namespace() string {
	return s.namespace
}

// Version returns the version of the service implementation.
func (s *simpleService) Version() string {
	return s.version
}

// Health returns the current health status of the service.
func (s *simpleService) Health() Health {
	status := HealthStatusUnknown
	if s.stopped.Load() {
		if s.err != nil {
			status = HealthStatusError
		} else {
			status = HealthStatusShutdown
		}
	} else if s.started.Load() {
		status = HealthStatusHealthy
	}

	return Health{
		Status: status,
		Error:  s.err,
	}
}

// Error returns the terminal error that caused the service to stop, if any.
// If the service is still running or has completed successfully, Error returns nil.
func (s *simpleService) Error() error {
	return s.err
}

// Initialize prepares the service for execution.
// For simpleService, this is a no-op and always returns nil.
func (s *simpleService) Initialize(_ *Context) error {
	return nil
}

// Run starts the main execution loop of the service.
// It ensures the service is only started once and not after it has been stopped.
func (s *simpleService) Run(ctx *Context) error {
	if s.started.Swap(true) {
		return ErrServiceAlreadyRunning
	}

	if s.stopped.Load() {
		return ErrServiceAlreadyStopped
	}

	defer func() {
		s.stopped.Store(true)
	}()

	// Check for shutdown request during execution
	if s.shutdownRequested.Load() {
		return nil
	}

	err := s.fn(ctx)
	if err != nil {
		s.err = err
	}
	return err
}

// Shutdown gracefully stops the service and releases all resources.
// For simpleService, this signals the service to stop on next opportunity.
func (s *simpleService) Shutdown(_ *Context) error {
	s.shutdownRequested.Store(true)
	return nil
}
