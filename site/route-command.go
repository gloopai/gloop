package site

import "sync"

// RouteCommandManager 管理路由命令的线程安全结构体
type RouteCommandManager struct {
	commands map[string]func(*RequestPayload) ResponsePayload
	mutex    sync.RWMutex
}

// NewRouteCommandManager 创建一个新的 RouteCommandManager
func NewRouteCommandManager() *RouteCommandManager {
	return &RouteCommandManager{
		commands: make(map[string]func(*RequestPayload) ResponsePayload),
	}
}

// Store 存储一个路由命令
func (rcm *RouteCommandManager) Store(key string, handler func(*RequestPayload) ResponsePayload) {
	if rcm == nil {
		panic("RouteCommandManager is nil")
	}
	rcm.mutex.Lock()
	defer rcm.mutex.Unlock()
	rcm.commands[key] = handler
}

// Load 加载一个路由命令
func (rcm *RouteCommandManager) Load(key string) (func(*RequestPayload) ResponsePayload, bool) {
	rcm.mutex.RLock()
	defer rcm.mutex.RUnlock()
	handler, ok := rcm.commands[key]
	return handler, ok
}
