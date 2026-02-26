package registry

import (
	"github.com/gloopai/gloop/modules"
	"github.com/gloopai/gloop/registry/consul"
)

type Registry struct {
	modules.Base
}

func NewRegistry(config consul.ConsulOptions) *Registry {
	return &Registry{}
}

func (r *Registry) Name() string {
	return "registry"
}

func (r *Registry) Init() {
	// 在这里可以添加初始化逻辑，例如注册服务等
}

func (r *Registry) Start() error {
	// 在这里可以添加启动逻辑，例如连接到服务发现系统等
	return nil
}

func (r *Registry) Close() {
	// 在这里可以添加关闭逻辑，例如断开与服务发现系统的连接等
}

func (r *Registry) Destroy() {
	// 在这里可以添加销毁逻辑，例如清理资源等
}
