package main

import (
	"context"

	"github.com/FlowSeer/fail"
	"github.com/FlowSeer/service"
)

func main() {
	svc := testService{}

	service.Run(context.Background(), svc)
}

type testService struct {
}

func (t testService) Name() string {
	return "test"
}

func (t testService) Namespace() string {
	return "test"
}

func (t testService) Version() string {
	return "0.0.1"
}

func (t testService) Error() error {
	return nil
}

func (t testService) Initialize(handle *service.Handle) error {
	handle.Logger().Info("test service initialized")
	return nil
}

func (t testService) Run(handle *service.Handle) error {
	handle.Logger().Info("test service running")
	return nil
}

func (t testService) Shutdown(handle *service.Handle) error {
	handle.Logger().Info("test service shutting down")
	return fail.Msg("test")
}
