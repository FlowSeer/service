package service

//go:generate go tool golang.org/x/tools/cmd/stringer -type Stage -trimprefix Stage

// Stage represents the current state of a service.
type Stage int

const (
	// StageWaiting indicates the service is waiting to start.
	StageWaiting Stage = iota
	// StageInitializing indicates the service is initializing.
	StageInitializing
	// StageRunning indicates the service is currently running.
	StageRunning
	// StageStopping indicates the service is in the process of stopping.
	StageStopping
	// StageFinished indicates the service has finished successfully.
	StageFinished
	// StageFailed indicates the service has failed.
	StageFailed
)
