package service

import "github.com/FlowSeer/fail"

var (
	// ErrServiceAlreadyRunning indicates that the service is already running.
	ErrServiceAlreadyRunning = fail.Msg("service is already running")
	// ErrServiceAlreadyStopped indicates that the service has already been run and stopped.
	ErrServiceAlreadyStopped = fail.Msg("service has already been run and is stopped")
	// ErrRunnerWaiting indicates that the runner is currently waiting for services to complete their execution.
	ErrRunnerWaiting = fail.Msg("runner is waiting for services to finish")
	// ErrRunnerAlreadyWaiting indicates that the runner is already waiting for services to complete their execution.
	ErrRunnerAlreadyWaiting = fail.Msg("runner is already waiting for services to finish")
	// ErrRunnerStopped indicates that the runner has been stopped.
	ErrRunnerStopped = fail.Msg("runner is stopped")
)
