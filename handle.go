package service

import (
	"context"
	"fmt"
	"sync"

	"github.com/FlowSeer/fail"
)

// Handle represents a managed or running service instance within the application.
// It provides methods to access the service's identity, current state, error status, and to initiate a graceful shutdown.
// Implementations of Handle are responsible for tracking the lifecycle and metadata of a service.
type Handle struct {
	// name is the unique name of the service.
	name string
	// namespace is the namespace to which the service belongs.
	namespace string
	// version is the version string of the service.
	version string
	// error is the last error encountered by the service, or nil if none.
	err    error
	errMtx sync.RWMutex
	// phase is the current lifecycle phase/state of the service.
	phase    Phase
	phaseMtx sync.RWMutex
	// exitSig is the channel singaling that the service has exited.
	// It is closed when the service has exited either successfully or due to an error.
	exitSig chan struct{}
	// shutdownFunc is the function to gracefully shut down the service.
	shutdownFunc func(context.Context) error
	shutdownOnce sync.Once

	shutdownErr    error
	shutdownErrMtx sync.RWMutex
}

func (h *Handle) String() string {
	if h.namespace != "" {
		return fmt.Sprintf("%s/%s @ %s", h.Namespace(), h.Name(), h.Version())
	}

	return fmt.Sprintf("%s @ %s", h.Name(), h.Version())
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
	return h.getError()
}

// Phase returns the current lifecycle phase or state of the service instance.
func (h *Handle) Phase() Phase {
	return h.getPhase()
}

// Wait blocks until the service has exited.
// It returns the last error encountered by the service, or nil if no error has occurred.
func (h *Handle) Wait() error {
	<-h.exitSig

	return fail.WithAssociated(h.Error(), h.getShutdownErr())
}

// Shutdown attempts to gracefully shut down the service instance, using the provided context for cancellation and timeout.
// Returns an error if shutdown fails or if the context is canceled or times out.
func (h *Handle) Shutdown(ctx context.Context) error {
	h.shutdownOnce.Do(func() {
		if err := h.shutdownFunc(ctx); err != nil {
			h.setShutdownErr(err)
		}
	})

	return h.Wait()
}

func (h *Handle) setStopped(err error) {
	h.setError(err)
	close(h.exitSig)
}

func (h *Handle) getPhase() Phase {
	h.phaseMtx.RLock()
	defer h.phaseMtx.RUnlock()

	return h.phase
}

func (h *Handle) setPhase(phase Phase) {
	h.phaseMtx.Lock()
	defer h.phaseMtx.Unlock()

	h.phase = phase
}

func (h *Handle) getError() error {
	h.errMtx.RLock()
	defer h.errMtx.RUnlock()

	return h.err
}

func (h *Handle) setError(err error) {
	h.errMtx.Lock()
	defer h.errMtx.Unlock()

	h.err = err
}

func (h *Handle) getShutdownErr() error {
	h.shutdownErrMtx.RLock()
	defer h.shutdownErrMtx.RUnlock()

	return h.shutdownErr
}

func (h *Handle) setShutdownErr(err error) {
	h.shutdownErrMtx.Lock()
	defer h.shutdownErrMtx.Unlock()

	h.shutdownErr = err
}

func createErrorHandle(svc Service, err error) *Handle {
	h := &Handle{
		name:      svc.Name(),
		namespace: svc.Namespace(),
		version:   svc.Version(),
		err:       err,
		exitSig:   make(chan struct{}),
	}
	// call with noop to forbid double-shutdown
	h.shutdownOnce.Do(func() {})
	close(h.exitSig)

	return h
}

func createHandle(svc Service, svcContext *Context) *Handle {
	return &Handle{
		name:      svc.Name(),
		namespace: svc.Namespace(),
		version:   svc.Version(),
		exitSig:   make(chan struct{}),
		shutdownFunc: func(ctx context.Context) error {
			return svc.Shutdown(svcContext)
		},
	}
}
