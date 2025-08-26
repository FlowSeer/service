package service

//go:generate go tool golang.org/x/tools/cmd/stringer -type Phase -trimprefix Phase

// Phase represents the current state of a service.
type Phase int

const (
	// PhaseWaiting indicates the service is waiting to start.
	PhaseWaiting Phase = iota
	// PhaseInitializing indicates the service is initializing.
	PhaseInitializing
	// PhaseRunning indicates the service is currently running.
	PhaseRunning
	// PhaseShuttingDown indicates the service is in the process of stopping.
	PhaseShuttingDown
	// PhaseFinished indicates the service has finished successfully.
	PhaseFinished
	// PhaseFailed indicates the service has failed.
	PhaseFailed
)
