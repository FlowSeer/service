package main

import (
	"time"

	"github.com/FlowSeer/fail"
	"github.com/FlowSeer/service"
)

const (
	serviceName      = "example"
	serviceVersion   = "0.0.1"
	serviceNamespace = "flowseer"
)

type exampleService struct{}

func (e exampleService) Name() string {
	return serviceName
}

func (e exampleService) Namespace() string {
	return serviceNamespace
}

func (e exampleService) Version() string {
	return serviceVersion
}

func (e exampleService) Health() service.Health {
	return service.Health{
		Status: service.HealthStatusUnknown,
	}
}

func (e exampleService) Error() error {
	return nil
}

func (e exampleService) Initialize(ctx *service.Context) error {
	type Config struct {
		Test string `json:"test"`
	}

	cfg, err := service.ReadConfig[Config](ctx)
	if err != nil {
		return err
	}

	ctx.Info("config: ", cfg)

	return nil
}

func (e exampleService) Run(ctx *service.Context) error {
	ctx.Info("Sleeping for 2 seconds...")
	time.Sleep(2 * time.Second)
	ctx.Info("Done sleeping.")
	return nil
}

func (e exampleService) Shutdown(ctx *service.Context) error {
	return fail.Msg("test")
}
