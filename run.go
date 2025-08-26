package service

import "context"

// Run starts the provided Service by invoking its Run method within the given context.
// This is a convenience wrapper that runs a single service using the same semantics as RunInParallel.
// It blocks until the service completes or the context is cancelled.
// The returned error is the result of the service's Run method or context cancellation.
func Run(ctx context.Context, svc Service) error {
	return RunInParallel(ctx, svc)
}

// RunInParallel starts all provided Services concurrently and waits for each to finish execution,
// regardless of whether any service fails or completes successfully.
//
// Each service's Run method is invoked in its own goroutine. If a service returns an error, it does not
// affect the execution of the other services; all services are allowed to run to completion or until their
// context is cancelled. The function returns only after all services have exited their Run methods.
//
// The returned error is an aggregate of all service errors. If none of the services fail, the error is nil.
// To inspect individual service errors, use the Error method on each Service instance after completion.
func RunInParallel(ctx context.Context, svcs ...Service) error {
	return run(ctx, false, svcs)
}

// RunInGroup starts all provided Services concurrently as a group and waits for them to finish.
// If any service's Run method returns (either due to error or normal completion), RunInGroup initiates
// a coordinated shutdown of all remaining services by cancelling their contexts and calling Shutdown.
//
// If the first service to exit did so due to an error, that error is returned. If other services encounter
// errors during their shutdown, those errors are attached as associated errors to the returned error.
// This function ensures that all services are stopped as soon as any one of them exits, providing a
// fail-fast, all-or-nothing group execution model.
func RunInGroup(ctx context.Context, svcs ...Service) error {
	return run(ctx, true, svcs)
}

// run is an internal helper function that manages the concurrent execution of one or more Service instances.
// It is used by both RunInParallel and RunInGroup to implement their respective service orchestration semantics.
func run(ctx context.Context, grouped bool, svcs []Service) error {
	panic("not implemented")
}
