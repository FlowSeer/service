package service

import (
	"context"

	"github.com/FlowSeer/fail"
	"go.opentelemetry.io/otel/metric"
	metricNoop "go.opentelemetry.io/otel/metric/noop"
	metricSdk "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	traceSdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/sync/errgroup"
)

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
	if len(svcs) == 0 {
		return nil
	}

	if len(svcs) == 1 {
		return runOne(ctx, svcs[0])
	}

	var eg = &errgroup.Group{}
	if grouped {
		eg, ctx = errgroup.WithContext(ctx)
	}

	for _, svc := range svcs {
		eg.Go(func() error {
			return runOne(ctx, svc)
		})
	}

	// FIXME: return the errors as documented
	return eg.Wait()
}

// runOne is an internal helper function that manages the execution of a single Service instance.
// It is used by both RunInParallel and RunInGroup to implement their respective service orchestration semantics.
func runOne(ctx context.Context, svc Service) error {
	ctx = fail.ContextWithAttributes(ctx, map[string]any{
		string(semconv.ServiceNameKey):      svc.Name(),
		string(semconv.ServiceVersionKey):   svc.Version(),
		string(semconv.ServiceNamespaceKey): svc.Namespace(),
	})

	logLevel := LogLevelFromEnv(svc.Name())
	logFormat := LogFormatFromEnv(svc.Name())
	logger := LoggerFromEnv(svc.Name()).
		With(
			string(semconv.ServiceNameKey), svc.Name(),
			string(semconv.ServiceVersionKey), svc.Version(),
			string(semconv.ServiceNamespaceKey), svc.Namespace(),
		)

	ctx = WithLogLevel(ctx, logLevel)
	ctx = WithLogFormat(ctx, logFormat)
	ctx = WithLogger(ctx, logger)

	var meterProvider metric.MeterProvider = metricNoop.NewMeterProvider()
	var tracerProvider trace.TracerProvider = traceSdk.NewTracerProvider()

	if IsOtelEnabled(svc.Name()) {
		res, err := resource.New(ctx, resource.WithAttributes(
			semconv.ServiceName(svc.Name()),
			semconv.ServiceVersion(svc.Version()),
		))
		if err != nil {
			return fail.New().
				Context(ctx).
				Cause(err).
				Msg("failed to create resource")
		}

		tracerProvider = TracerProviderFromEnv(traceSdk.WithResource(res))
		ctx = WithTracerProvider(ctx, tracerProvider)

		meterProvider = MeterProviderFromEnv(metricSdk.WithResource(res))
		ctx = WithMeterProvider(ctx, meterProvider)
	}

	handle := &Handle{
		ctx:            ctx,
		meterProvider:  meterProvider,
		tracerProvider: tracerProvider,
	}

	handle.phase = PhaseInitializing
	handle.logger = logger.With(
		"service.phase", handle.phase.String(),
	)
	if err := svc.Initialize(handle); err != nil {
		return fail.New().
			Context(ctx).
			Cause(err).
			Msg("failed to initialize service")
	}

	handle.phase = PhaseRunning
	handle.logger = logger.With(
		"service.phase", handle.phase.String(),
	)

	err := svc.Run(handle)

	handle.phase = PhaseShuttingDown
	handle.logger = logger.With(
		"service.phase", handle.phase.String(),
	)
	shutdownErr := svc.Shutdown(handle)

	if err != nil {
		return fail.New().
			Context(ctx).
			Cause(err).
			Associate(shutdownErr).
			Msg("service failed")
	}

	return fail.Wrap("serviced failed to shutdown cleanly", shutdownErr)
}
