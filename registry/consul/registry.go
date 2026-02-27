package consul

import (
	"fmt"

	"github.com/gloopai/gloop/lib"
	"github.com/hashicorp/consul/api"
)

type Registry struct {
	config      *Options
	serviceId   string
	serviceName string
	client      *api.Client
}

func NewRegistry(ops *Options) *Registry {
	config := api.DefaultConfig()
	config.Address = lib.Conf.GetString(ops.Addr, "127.0.0.1:8500")
	client, _ := api.NewClient(config)

	return &Registry{
		config: ops,
		client: client,
	}
}

// 注册服务
func (r *Registry) Register(serviceID, serviceName, serviceAddr string, servicePort int) error {
	r.serviceId = serviceID
	r.serviceName = serviceName
	registration := &api.AgentServiceRegistration{
		ID:      r.serviceId,
		Name:    r.serviceName,
		Port:    servicePort,
		Address: serviceAddr,
		Check: &api.AgentServiceCheck{
			GRPC:     fmt.Sprintf("%s", serviceAddr),
			Interval: "10s",
			Timeout:  "5s",
		},
	}
	return r.client.Agent().ServiceRegister(registration)
}

// 注消服务
func (r *Registry) Close() {
	r.client.Agent().ServiceDeregister(r.serviceId)
}

// 获取服务链接
func (r *Registry) GetServiceTarget(serviceName string) (string, error) {
	target := fmt.Sprintf("consul://%s/%s?wait=14s", r.config.Addr, serviceName)
	return target, nil
}
