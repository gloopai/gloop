package node

import (
	"context"
	"errors"
	"net/http"

	"github.com/gloopai/gloop/events"
	"github.com/gloopai/gloop/modules/auth"
	"github.com/gloopai/gloop/modules/pkg"
	"github.com/gloopai/gloop/schema"
)

type Proxy struct {
	NodeId   string
	NodeName string
	Node     *Node
}

// AddRestRouter 添加 REST 路由
func (p *Proxy) RestAddRoute(pattern string, handlerFunc http.HandlerFunc) {
	p.Node.rest.AddRoute(pattern, handlerFunc)
}

// AddRestPayloadRoute 添加处理 JSON 请求体的 REST 路由
func (p *Proxy) RestAddPayloadRoute(pattern string) {
	p.Node.rest.AddPayloadRoute(pattern)
}

// AddRestRouterWithAuth 添加带认证的 REST 路由
func (p *Proxy) RestAddPayloadRouteWithAuth(pattern string) {
	p.Node.rest.AddPayloadRouteWithAuth(pattern)
}

// RegisterRestRouteCommand 注册 REST 路由命令
func (p *Proxy) RestRegisterRouteCommand(route string, command string, handler func(ctx context.Context, payload *schema.Request) schema.Response) {
	p.Node.rest.RegisterCommand(route, command, handler)
}

// GetRestAuth 获取 REST 模块的认证组件
func (p *Proxy) RestGetAuth() *auth.Auth {
	return p.Node.rest.Auth
}

// EventTrigger 触发事件
func (p *Proxy) EventTrigger(eventName string, data interface{}) {
	p.Node.events.Publish(eventName, data)
}

// EventSyncTrigger 同步触发事件
func (p *Proxy) EventSyncTrigger(eventName string, data interface{}) {
	p.Node.events.SyncPublish(eventName, data)
}

// EventOn 订阅事件
func (p *Proxy) EventOn(eventName string, handler events.EventHandler) {
	p.Node.events.Subscribe(eventName, handler)
}

// / 获取事件总线
func (p *Proxy) GetEventBus() *events.EventBus {
	return p.Node.events
}

// GetMysql 获取 MySQL 客户端
func (p *Proxy) GetMysql() (*pkg.MysqlClient, error) {
	if p.Node.mysql == nil {
		return nil, errors.New("MySQL client is not initialized")
	}
	return p.Node.mysql, nil
}
