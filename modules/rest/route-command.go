package rest

import (
	"context"
	"sync"

	"github.com/gloopai/gloop/schema"
)

// RouteCommandManager 管理路由命令的线程安全结构体
type RouteCommandManager struct {
	commands map[string]func(ctx context.Context, payload *schema.Request) schema.Response
	mutex    sync.RWMutex
}

// NewRouteCommandManager 创建一个新的 RouteCommandManager
func NewRouteCommandManager() *RouteCommandManager {
	return &RouteCommandManager{
		commands: make(map[string]func(ctx context.Context, payload *schema.Request) schema.Response),
	}
}

// Store 存储一个路由命令
func (rcm *RouteCommandManager) Store(key string, handler func(ctx context.Context, payload *schema.Request) schema.Response) {
	if rcm == nil {
		panic("RouteCommandManager is nil")
	}
	rcm.mutex.Lock()
	defer rcm.mutex.Unlock()
	rcm.commands[key] = handler
}

// Load 加载一个路由命令
func (rcm *RouteCommandManager) Load(key string) (func(ctx context.Context, payload *schema.Request) schema.Response, bool) {
	rcm.mutex.RLock()
	defer rcm.mutex.RUnlock()
	handler, ok := rcm.commands[key]
	return handler, ok
}
