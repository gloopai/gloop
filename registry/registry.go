package registry

import (
	"github.com/gloopai/gloop/registry/consul"
)

type RegistryOptions struct {
	Addr string
}

type Registry struct {
}

func NewRegistry(config consul.ConsulOptions) *Registry {
	return &Registry{}
}
