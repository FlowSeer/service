package service

import (
	"context"
)

// Handle represents a managed or running service instance within the application.
// It provides methods to access the service's identity, current state, error status, and to initiate a graceful shutdown.
// Implementations of Handle are responsible for tracking the lifecycle and metadata of a service.
type Handle struct {
	id        string
	name      string                      // The unique name of the service.
	namespace string                      // The namespace to which the service belongs.
	version   string                      // The version string of the service.
	error     error                       // The last error encountered by the service, or nil if none.
	phase     Phase                       // The current lifecycle phase/state of the service.
	shutdown  func(context.Context) error // Function to gracefully shut down the service.
}

// Id returns the unique identifier of the service instance.
func (h *Handle) Id() string {
	return h.id
}

// Name returns the unique name of the service instance.
func (h *Handle) Name() string {
	return h.name
}

// Namespace returns the namespace associated with the service instance.
func (h *Handle) Namespace() string {
	return h.namespace
}

// Version returns the version string of the service instance.
func (h *Handle) Version() string {
	return h.version
}

// Error returns the most recent error encountered by the service, or nil if no error has occurred.
func (h *Handle) Error() error {
	return h.error
}

// Phase returns the current lifecycle phase or state of the service instance.
func (h *Handle) Phase() Phase {
	return h.phase
}

// Shutdown attempts to gracefully shut down the service instance, using the provided context for cancellation and timeout.
// Returns an error if shutdown fails or if the context is canceled or times out.
func (h *Handle) Shutdown(ctx context.Context) error {
	return h.shutdown(ctx)
}
