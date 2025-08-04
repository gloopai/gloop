package servicehub

import (
	"sync"

	"github.com/gloopai/gloop/lib"
)

var (
	singleton     *ServiceHub
	singletonOnce sync.Once
)

// GetHubInstance 获取 ServiceHub 单例
func GetHubInstance() *ServiceHub {
	singletonOnce.Do(func() {
		singleton = NewServiceHub()
	})
	return singleton
}

// ServiceFunc 定义服务函数的类型，使用泛型提升类型安全
type ServiceFunc[Req any, Resp any] func(req Req) (Resp, error)

// ServiceHub 用于模块间注册和调用服务
// 支持并发安全

// ServiceHub 支持不同类型服务的注册和调用
type serviceEntry struct {
	fn          any
	description string
}

type ServiceHub struct {
	services map[string]serviceEntry
	mu       sync.RWMutex
}

// NewServiceHub 创建一个新的 ServiceHub 实例
func NewServiceHub() *ServiceHub {
	return &ServiceHub{
		services: make(map[string]serviceEntry),
	}
}

// Register 注册一个服务，支持覆盖和描述
func Register[Req any, Resp any](h *ServiceHub, name string, fn ServiceFunc[Req, Resp], opts ...RegisterOption) error {
	lib.Log.Debugf("\033[33m[ServiceHub] register service %s\033[0m", name)
	h.mu.Lock()
	defer h.mu.Unlock()
	var opt registerOptions
	for _, o := range opts {
		o(&opt)
	}
	if _, exists := h.services[name]; exists && !opt.allowOverride {
		return &ErrServiceExists{name}
	}
	h.services[name] = serviceEntry{fn: fn, description: opt.description}
	return nil
}

type registerOptions struct {
	allowOverride bool
	description   string
}
type RegisterOption func(*registerOptions)

// WithOverride 允许覆盖已注册服务
func WithOverride() RegisterOption {
	return func(o *registerOptions) { o.allowOverride = true }
}

// WithDescription 添加服务描述
func WithDescription(desc string) RegisterOption {
	return func(o *registerOptions) { o.description = desc }
}

// Call 调用一个已注册的服务，返回详细错误（带彩色日志）
func Call[Req any, Resp any](h *ServiceHub, name string, req Req) (Resp, error) {
	// 蓝色: \033[34m，重置: \033[0m
	lib.Log.Debugf("\033[33m[ServiceHub] call service %s\033[0m", name)
	h.mu.RLock()
	entry, exists := h.services[name]
	h.mu.RUnlock()
	if !exists {
		var zero Resp
		// 红色: \033[31m
		lib.Log.Error("service not found!!", "name", name)
		return zero, &ErrServiceNotFound{name}
	}
	typedFn, ok := entry.fn.(ServiceFunc[Req, Resp])
	if !ok {
		var zero Resp
		lib.Log.Error("mservice type mismatch!!", "name", name)
		return zero, &ErrServiceType{name}
	}
	return typedFn(req)
}

// Unregister 注销一个服务
func (h *ServiceHub) Unregister(name string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.services, name)
}

// UnregisterMany 批量注销服务
func (h *ServiceHub) UnregisterMany(names ...string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for _, name := range names {
		delete(h.services, name)
	}
}

// Has 检查服务是否存在
func (h *ServiceHub) Has(name string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	_, exists := h.services[name]
	return exists
}

// ListServices 获取所有已注册服务名
func (h *ServiceHub) ListServices() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	names := make([]string, 0, len(h.services))
	for name := range h.services {
		names = append(names, name)
	}
	return names
}

// GetServiceDescription 获取服务描述
func (h *ServiceHub) GetServiceDescription(name string) (string, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	entry, exists := h.services[name]
	if !exists {
		return "", false
	}
	return entry.description, true
}

// 错误类型
type ErrServiceExists struct{ Name string }

func (e *ErrServiceExists) Error() string { return "service '" + e.Name + "' already registered" }

type ErrServiceNotFound struct{ Name string }

func (e *ErrServiceNotFound) Error() string { return "service '" + e.Name + "' not found" }

type ErrServiceType struct{ Name string }

func (e *ErrServiceType) Error() string { return "service '" + e.Name + "' type mismatch" }

// RegisterToService 注册一个服务服到 HUB 单例
func RegisterToService[Req any, Resp any](name string, fn ServiceFunc[Req, Resp], opts ...RegisterOption) error {
	return Register(GetHubInstance(), name, fn, opts...)
}

// CallFromService 通过单例调用一个服务
func CallFromService[Req any, Resp any](name string, req Req) (Resp, error) {
	return Call[Req, Resp](GetHubInstance(), name, req)
}
