package service

import "github.com/FlowSeer/fail"

var (
	// ErrRunnerWaiting indicates that the runner is currently waiting for services to complete their execution.
	ErrRunnerWaiting = fail.Msg("runner is waiting for services to finish")
	// ErrRunnerAlreadyWaiting indicates that the runner is already waiting for services to complete their execution.
	ErrRunnerAlreadyWaiting = fail.Msg("runner is already waiting for services to finish")
	// ErrRunnerStopped indicates that the runner has been stopped.
	ErrRunnerStopped = fail.Msg("runner is stopped")
)
