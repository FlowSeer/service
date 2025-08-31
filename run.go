package service

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/FlowSeer/fail"
	"go.opentelemetry.io/contrib/instrumentation/host"
	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel/metric"
	metricNoop "go.opentelemetry.io/otel/metric/noop"
	metricSdk "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	traceSdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
	traceNoop "go.opentelemetry.io/otel/trace/noop"
	"golang.org/x/sync/errgroup"
)

// RunAndExit runs the given service using the provided context, waits for it to finish,
// and then exits the process with an appropriate exit code based on the error returned.
// If the service completes successfully, the process exits with code 0.
// If an error occurs, the process exits with the code returned by fail.ExitCode(err).
func RunAndExit(ctx context.Context, svc Service) {
	err := RunAndWait(ctx, svc)
	if err != nil {
		println(fail.PrintPretty(err))
		os.Exit(fail.ExitCode(err))
	} else {
		os.Exit(0)
	}
}

// RunParallelAndExit runs multiple services in parallel using the provided context,
// waits for all of them to finish, and then exits the process with the highest exit code
// among all returned errors. If all services complete successfully, the process exits with code 0.
func RunParallelAndExit(ctx context.Context, svcs ...Service) {
	errs := RunParallelAndWait(ctx, svcs...)

	exitCode := 0
	for _, err := range errs {
		exitCode = max(exitCode, fail.ExitCode(err))
	}

	os.Exit(exitCode)
}

// RunGroupAndExit runs multiple services as a group using the provided context,
// where the group is canceled if any service returns an error. It waits for all services
// to finish and then exits the process with the highest exit code among all returned errors.
// If all services complete successfully, the process exits with code 0.
func RunGroupAndExit(ctx context.Context, svcs ...Service) {
	errs := RunGroupAndWait(ctx, svcs...)

	exitCode := 0
	for _, err := range errs {
		exitCode = max(exitCode, fail.ExitCode(err))
	}

	os.Exit(exitCode)
}

// RunAndWait runs the given service using the provided context and waits for it to finish.
// It returns the error returned by the service, or nil if the service completes successfully.
func RunAndWait(ctx context.Context, svc Service) error {
	return Run(ctx, svc).Wait()
}

// RunParallelAndWait runs multiple services in parallel using the provided context,
// waits for all of them to finish, and returns a slice of errors corresponding to each service.
// If a service completes successfully, its error will be nil.
func RunParallelAndWait(ctx context.Context, svcs ...Service) []error {
	switch len(svcs) {
	case 0:
		return nil
	case 1:
		return []error{RunAndWait(ctx, svcs[0])}
	}

	wg := sync.WaitGroup{}
	handles := RunParallel(ctx, svcs...)
	errs := make([]error, len(handles))
	for i, h := range handles {
		wg.Add(1)

		go func(h *Handle) {
			defer wg.Done()
			errs[i] = h.Wait()
		}(h)
	}

	wg.Wait()
	return errs
}

// RunGroupAndWait runs multiple services as a group using the provided context,
// where the group is canceled if any service returns an error. It waits for all services
// to finish and returns a slice of errors corresponding to each service.
// If a service completes successfully, its error will be nil.
func RunGroupAndWait(ctx context.Context, svcs ...Service) []error {
	switch len(svcs) {
	case 0:
		return nil
	case 1:
		return []error{RunAndWait(ctx, svcs[0])}
	}

	wg := sync.WaitGroup{}
	handles := RunGroup(ctx, svcs...)
	errs := make([]error, len(handles))
	for i, h := range handles {
		wg.Add(1)

		go func(h *Handle) {
			defer wg.Done()
			errs[i] = h.Wait()
		}(h)
	}

	wg.Wait()
	return errs
}

// Run runs the given service using the provided context and returns a Handle
// that can be used to wait for the service to finish or to shut it down.
func Run(ctx context.Context, svc Service) *Handle {
	return RunParallel(ctx, svc)[0]
}

// RunParallel runs multiple services in parallel using the provided context and returns
// a slice of Handles, one for each service. The services are run independently and are not
// canceled if any other service fails.
func RunParallel(ctx context.Context, svcs ...Service) []*Handle {
	return runAll(ctx, false, svcs)
}

// RunGroup runs multiple services as a group using the provided context and returns
// a slice of Handles, one for each service. If any service returns an error, the context
// is canceled for all services in the group.
func RunGroup(ctx context.Context, svcs ...Service) []*Handle {
	return runAll(ctx, true, svcs)
}

// runAll runs the services using the provided context and error group.
// if any service returns an error. Returns a slice of Handles for the running services.
func runAll(ctx context.Context, grouped bool, svcs []Service) []*Handle {
	eg := &errgroup.Group{} // empty group is valid and implies no cancellation on error
	if grouped {
		eg, ctx = errgroup.WithContext(ctx)
	}

	handles := make([]*Handle, len(svcs))
	for i, svc := range svcs {
		handles[i] = run(ctx, eg, svc)
	}

	return handles
}

// run runs the given service using the provided context and returns a Handle
// that can be used to wait for the service to finish or to shut it down.
// The service is being run in parallel using the provided error group.
func run(ctx context.Context, eg *errgroup.Group, svc Service) *Handle {
	svcCtx, err := createContext(ctx, svc)
	if err != nil {
		return createErrorHandle(svc, err)
	}

	handle := createHandle(svc, svcCtx)
	eg.Go(func() error {
		svcErr := runBlocking(svcCtx, svc, handle)
		handle.setStopped(svcErr)

		return svcErr
	})

	return handle
}

func runBlocking(ctx *Context, svc Service, handle *Handle) error {
	ctx.Logger().Debug("Initializing")
	handle.setPhase(PhaseInitializing)

	err := svc.Initialize(ctx)
	if err != nil {
		return err
	}

	ctx.Logger().Debug("Running")
	handle.setPhase(PhaseRunning)

	err = svc.Run(ctx)
	if err != nil {
		return err
	}

	ctx.Logger().Debug("Shutting down")
	handle.setPhase(PhaseShuttingDown)

	shutdownErr := handle.Shutdown(ctx)
	if shutdownErr != nil {
		handle.setPhase(PhaseFailed)
	} else {
		handle.setPhase(PhaseFinished)
	}

	if err != nil {
		return fail.WithAssociated(err, shutdownErr)
	} else {
		return shutdownErr
	}
}

func createContext(ctx context.Context, svc Service) (*Context, error) {
	ctx = fail.ContextWithAttributes(ctx, map[string]any{
		"service.name":      svc.Name(),
		"service.version":   svc.Version(),
		"service.namespace": svc.Namespace(),
	})

	logger := LoggerFromEnv(svc.Name()).
		With("service.name", svc.Name(),
			"service.version", svc.Version())
	if svc.Namespace() != "" {
		logger = logger.With("service.namespace", svc.Namespace())
	}

	ctx = WithLogger(ctx, logger)

	var (
		tracerProvider trace.TracerProvider
		tracerShutdown OtelShutdownFunc
		meterProvider  metric.MeterProvider
		meterShutdown  OtelShutdownFunc
	)
	if IsOtelEnabled(svc.Name()) {
		res, err := resource.New(ctx, resource.WithAttributes(
			semconv.ServiceName(svc.Name()),
			semconv.ServiceVersion(svc.Version()),
			semconv.ServiceNamespace(svc.Namespace()),
		))
		if err != nil {
			return nil, fail.Wrap(err, "failed to create OTEL resource")
		}

		tracerProvider, tracerShutdown, err = TracerProviderFromEnv(ctx, traceSdk.WithResource(res))
		if err != nil {
			return nil, fail.Wrap(err, "failed to create OTEL tracer provider")
		}

		meterProvider, meterShutdown, err = MeterProviderFromEnv(ctx, metricSdk.WithResource(res))
		if err != nil {
			return nil, fail.Wrap(err, "failed to create OTEL meter provider")
		}

		err = runtime.Start(runtime.WithMeterProvider(meterProvider))
		if err != nil {
			return nil, fail.Wrap(err, "failed to start collection of runtime metrics")
		}

		err = host.Start(host.WithMeterProvider(meterProvider))
		if err != nil {
			return nil, fail.Wrap(err, "failed to start collection of host metrics")
		}
	} else {
		logger.Warn(fmt.Sprintf(
			"Set env %s=true to enable OpenTelemetry.",
			EnvName(svc.Name(), OtelEnableEnvVar)),
		)

		tracerProvider = traceNoop.NewTracerProvider()
		tracerShutdown = OtelNoopShutdown
		meterProvider = metricNoop.NewMeterProvider()
		meterShutdown = OtelNoopShutdown
	}

	ctx = WithTracerProvider(ctx, tracerProvider)
	ctx = WithMeterProvider(ctx, meterProvider)

	tracer := tracerProvider.Tracer(InstrumentationName, trace.WithInstrumentationVersion(InstrumentationVersion))
	ctx = WithTracer(ctx, tracer)

	meter := meterProvider.Meter(InstrumentationName, metric.WithInstrumentationVersion(InstrumentationVersion))
	ctx = WithMeter(ctx, meter)

	return &Context{
		Context:        ctx,
		logger:         logger,
		tracerProvider: tracerProvider,
		tracerShutdown: tracerShutdown,
		defaultTracer:  tracer,
		meterProvider:  meterProvider,
		meterShutdown:  meterShutdown,
		defaultMeter:   meter,
	}, nil
}
