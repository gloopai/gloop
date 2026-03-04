package gate

import (
	"errors"

	"github.com/gloopai/gloop/events"
	"github.com/gloopai/gloop/modules/pkg"
	"google.golang.org/grpc"
)

type Proxy struct {
	NodeId   string
	NodeName string
	Node     *Node
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

// GetRedis 获取 Redis 客户端
func (p *Proxy) GetRedis() (*pkg.RedisClient, error) {
	if p.Node.rdb == nil {
		return nil, errors.New("Redis client is not initialized")
	}
	return p.Node.rdb, nil
}

// 注册 gRpc 服务
func (p *Proxy) AddServiceProvider(desc *grpc.ServiceDesc, provider any) {
	p.Node.transporter.AddServiceProvider(desc, provider)
}

// 获取 gRPC 客户端
func (p *Proxy) GetServiceClient() (*grpc.ClientConn, error) {
	return p.Node.ServiceClient()
}
