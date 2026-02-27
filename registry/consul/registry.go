package consul

import (
	"fmt"
	"net"
	"strconv"

	"github.com/gloopai/gloop/lib"
	"github.com/hashicorp/consul/api"
)

type Registry struct {
	config      *Options
	serviceId   string
	serviceName string
	client      *api.Client
}

func NewRegistry(ops *Options) (*Registry, error) {
	config := api.DefaultConfig()
	config.Address = lib.Conf.GetString(ops.Addr, "127.0.0.1:8500")
	client, err := api.NewClient(config)
	if err != nil {
		return nil, err
	}

	return &Registry{
		config: ops,
		client: client,
	}, nil
}

// 注册服务
func (r *Registry) Register(serviceID, serviceName string, serviceAddr string) error {
	r.serviceId = serviceID
	r.serviceName = serviceName
	addr := ""
	port := 0
	host, portStr, err := net.SplitHostPort(serviceAddr)
	if err == nil {
		addr = host
		if p, err2 := strconv.Atoi(portStr); err2 == nil {
			port = p
		}
	} else {
		// fallback: keep whole string as address
		addr = serviceAddr
	}

	registration := &api.AgentServiceRegistration{
		ID:      r.serviceId,
		Name:    r.serviceName,
		Address: addr,
		Port:    port,
		Check: &api.AgentServiceCheck{
			GRPC:     fmt.Sprintf("%s", serviceAddr),
			Interval: "10s",
			Timeout:  "5s",
		},
	}

	return r.client.Agent().ServiceRegister(registration)
}

// 注消服务
func (r *Registry) Close() error {
	return r.client.Agent().ServiceDeregister(r.serviceId)
}

// 获取服务链接
func (r *Registry) GetServiceTarget(serviceName string) (string, error) {
	addr := ""
	if addr == "" && r.config != nil {
		addr = lib.Conf.GetString(r.config.Addr, "127.0.0.1:8500")
	}
	target := fmt.Sprintf("consul://%s/%s?wait=14s", addr, serviceName)
	return target, nil
}
