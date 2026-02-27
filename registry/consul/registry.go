package consul

import (
	"fmt"

	"github.com/hashicorp/consul/api"
)

type Registry struct {
}

func NewRegistry(ops *Options) *Registry {
	port := 50051
	serviceName := "hello-service"
	instanceID := "hello-service-1"

	// 1. 注册到 Consul
	config := api.DefaultConfig()
	config.Address = "127.0.0.1:8500"
	client, _ := api.NewClient(config)

	registration := &api.AgentServiceRegistration{
		ID:      instanceID,
		Name:    serviceName,
		Port:    port,
		Address: "127.0.0.1",
		Check: &api.AgentServiceCheck{
			// Consul 官方支持 gRPC 健康检查
			GRPC:     fmt.Sprintf("127.0.0.1:%d", port),
			Interval: "10s",
			Timeout:  "5s",
		},
	}
	client.Agent().ServiceRegister(registration)
	return &Registry{}
}
