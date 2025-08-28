package service

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/FlowSeer/fail"
)

// Runner manages the lifecycle of one or more Service instances.
// It provides methods to start, shut down, and wait for services, either individually or collectively.
// Implementations of Runner are responsible for orchestrating service execution, graceful shutdown, and error handling.
type Runner interface {
	// Run starts the given Service and returns a Handle for the running service,
	// along with any error encountered during startup.
	// The same service may be started multiple times, and the service implementation must make sure to handle this.
	Run(svc Service) (*Handle, error)

	// Shutdown attempts to gracefully shut down the specified service Handle using the provided context.
	// Returns an error if shutdown fails or if the context is canceled or times out.
	Shutdown(ctx context.Context, h *Handle) error

	// Stop running any new service and shut down all already-running services.
	Stop(ctx context.Context) error

	// Wait blocks until all managed services have fully stopped or the context is canceled.
	// The runner must not accept new service requests after a call to Wait has been made.
	// Returns an error if waiting fails or the context is canceled or times out.
	Wait(ctx context.Context) error
}

// RunnerOptions configures the behavior of a Runner.
type RunnerOptions struct {
	// AllOrNothing determines whether to abort all services if any service fails.
	// If true, the runner will immediately return an error if any service fails.
	// If false, the runner will continue running all services and return an error only after all services have exited.
	AllOrNothing bool
	// ExitOnError determines whether to exit the process if any service fails.
	ExitOnError bool
}

// RunnerOption is a function that modifies RunnerOptions.
// It is used to configure optional behaviors for a Runner at creation time.
type RunnerOption func(*RunnerOptions)

// WithAllOrNothing returns a RunnerOption that sets the AllOrNothing field of RunnerOptions.
// If allOrNothing is true, the runner will abort all services if any service fails.
// If false, the runner will allow all services to run to completion, even if some fail.
func WithAllOrNothing() RunnerOption {
	return func(opts *RunnerOptions) {
		opts.AllOrNothing = true
	}
}

// WithExitOnError returns a RunnerOption that sets the ExitOnError field of RunnerOptions.
// If exitOnError is true, the runner will exit the process if it fails.
// If AllOrNothing is true, the runner will exit the process if any service fails. Otherwise,
// the runner will continue running all services and exit the process only after all services have exited.
// The exit code is determined by the fail.ExitCode function.
func WithExitOnError() RunnerOption {
	return func(opts *RunnerOptions) {
		opts.ExitOnError = true
	}
}

// Run starts the provided Service by invoking its Run method within the given context.
// This is a convenience wrapper that runs a single service using the same semantics as RunAll.
// It blocks until the service completes or the context is canceled.
// The returned error is the result of the service's Run method or context cancellation.
func Run(ctx context.Context, svc Service, opts ...RunnerOption) error {
	return RunAll(ctx, []Service{svc}, opts...)
}

// RunAll starts all provided services using a new Runner configured with the given options.
// It blocks until all services complete or the context is canceled.
// Returns an error if any service fails to start or if waiting fails.
func RunAll(ctx context.Context, svcs []Service, opts ...RunnerOption) error {
	runner := NewRunner(ctx, opts...)

	for i, svc := range svcs {
		_, err := runner.Run(svc)

		if err != nil {
			// If we already added services, we need to shut them down before returning.
			if i > 0 {
				err = fail.WithAssociated(err, runner.Stop(ctx))
			}

			return err
		}
	}

	return runner.Wait(ctx)
}

// NewRunner creates a new DefaultRunner with the provided context and options.
func NewRunner(ctx context.Context, opts ...RunnerOption) Runner {
	options := RunnerOptions{}

	for _, opt := range opts {
		opt(&options)
	}

	return &DefaultRunner{
		opts:           options,
		ctx:            ctx,
		serviceHandles: make(map[string]*Handle),
		services:       make(map[string]Service),
	}
}

// DefaultRunner is a basic implementation of the Runner interface.
type DefaultRunner struct {
	opts RunnerOptions
	ctx  context.Context

	services       map[string]Service
	serviceHandles map[string]*Handle
	servicesMtx    sync.RWMutex

	stopping atomic.Bool
}

func (r *DefaultRunner) Run(svc Service) (*Handle, error) {
	handle := &Handle{
		id:        fmt.Sprintf("%s/%s@%s-%d", svc.Namespace(), svc.Name(), svc.Version(), time.Now().UnixNano()),
		name:      svc.Name(),
		namespace: svc.Namespace(),
		version:   svc.Version(),
		phase:     PhaseWaiting,
	}
	handle.shutdown = func(ctx context.Context) error {
		return r.Shutdown(ctx, handle)
	}

	r.servicesMtx.Lock()
	r.services[handle.id] = svc
	r.serviceHandles[handle.id] = handle
	r.servicesMtx.Unlock()

	return handle, r.run(r.ctx, svc, handle)
}

func (r *DefaultRunner) Shutdown(ctx context.Context, h *Handle) error {
	return r.shutdownAndRemove(ctx, h.Id())
}

func (r *DefaultRunner) Stop(ctx context.Context) error {
	if r.stopping.Swap(true) {
		return ErrServiceAlreadyStopped
	}

	return nil
}

func (r *DefaultRunner) Wait(ctx context.Context) error {
	panic("not implemented")
}

func (r *DefaultRunner) run(ctx context.Context, svc Service, handle *Handle) error {
	panic("not implemented")
}

func (r *DefaultRunner) shutdownAndRemove(ctx context.Context, id string) error {
	r.servicesMtx.RLock()
	handle := r.serviceHandles[id]
	r.servicesMtx.RUnlock()

	if handle == nil {
		return nil
	}

	err := r.Shutdown(ctx, handle)

	r.servicesMtx.Lock()
	delete(r.serviceHandles, id)
	delete(r.services, id)
	r.servicesMtx.Unlock()

	return err
}
