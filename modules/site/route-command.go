package site

import (
	"sync"

	"github.com/gloopai/gloop/modules"
)

// RouteCommandManager 管理路由命令的线程安全结构体
type RouteCommandManager struct {
	commands map[string]func(*modules.RequestPayload) modules.ResponsePayload
	mutex    sync.RWMutex
}

// NewRouteCommandManager 创建一个新的 RouteCommandManager
func NewRouteCommandManager() *RouteCommandManager {
	return &RouteCommandManager{
		commands: make(map[string]func(*modules.RequestPayload) modules.ResponsePayload),
	}
}

// Store 存储一个路由命令
func (rcm *RouteCommandManager) Store(key string, handler func(*modules.RequestPayload) modules.ResponsePayload) {
	if rcm == nil {
		panic("RouteCommandManager is nil")
	}
	rcm.mutex.Lock()
	defer rcm.mutex.Unlock()
	rcm.commands[key] = handler
}

// Load 加载一个路由命令
func (rcm *RouteCommandManager) Load(key string) (func(*modules.RequestPayload) modules.ResponsePayload, bool) {
	rcm.mutex.RLock()
	defer rcm.mutex.RUnlock()
	handler, ok := rcm.commands[key]
	return handler, ok
}
