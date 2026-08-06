package main

import (
	"context"
	arupa "github.com/SteelDrEgg/arupa-sdk/golang"
	"github.com/SteelDrEgg/arupa-sdk/golang/wasm"
)

const (
	serviceName = "get-started"
)

type myService struct {
	runtime *wasm.Service
	logger  arupa.Logger
}

func newService() *myService {
	service := &myService{}
	service.runtime = &wasm.Service{
		Info:       arupa.ServiceInfo{Name: serviceName, Version: serviceVersion},
		OnRegister: service.configure,
	}
	return service
}

func (s *myService) configure(ctx context.Context) error {
	if s.logger == nil {
		return nil
	}
	_ = s.logger.Log(ctx, arupa.LogInfo, "Registration completed")
	return nil
}
