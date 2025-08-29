package main

import (
	"context"

	"github.com/FlowSeer/service"
)

func main() {
	service.RunAndExit(context.Background(), exampleService{})
}
