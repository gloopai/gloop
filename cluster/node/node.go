package node

import (
	"fmt"

	"github.com/gloopai/gloop/events"
	"github.com/gloopai/gloop/lib"
	"github.com/gloopai/gloop/modules/pkg"
	"github.com/gloopai/gloop/modules/rest"
	"github.com/gloopai/gloop/registry/consul"
	ggrpc "github.com/gloopai/gloop/transport/grpc"
	"google.golang.org/grpc"
)

// Node 组件
type Node struct {
	NodeId      string
	NodeName    string
	Config      *NodeOptions
	events      *events.EventBus
	transporter *ggrpc.Transporter
	registry    *consul.Registry
	rdb         *pkg.RedisClient
	mysql       *pkg.MysqlClient
	nsq         *pkg.NsqClient
	rest        *rest.Rest
}

type NodeOptions struct {
	// 节点 id，全局必须唯一
	Id string
	// 节点名称，便于识别
	Name string
	// 节点grpc 地址，格式为 "ip:port"，如果没设置就随机端口
	Addr string
	// 注册中心配置
	Consul consul.Options
	// Redis 配置
	Redis pkg.RedisClientOptions
	// MySQL 配置
	Mysql pkg.MysqlClientOptions
	// Nsq 配置
	Nsq pkg.NsqClientOptions
	// Rest 配置
	Rest rest.RestOptions
}

func NewNode(config *NodeOptions) *Node {
	node := &Node{
		Config: config,
	}
	node.NodeId = lib.Conf.GetString(config.Id, lib.Generate.Guid())
	node.NodeName = lib.Conf.GetString(config.Name, "node")
	// 初始化 GRpc transporter
	transporter, err := ggrpc.NewTransporter(&ggrpc.Options{
		Addr: node.Config.Addr,
	})
	if err != nil {
		lib.Log.Fatalf("failed to create gRPC transport: %v", err)
	}
	node.transporter = transporter

	// 初始化 MySQL 客户端
	if node.Config.Mysql.DSN != "" {
		mysqlClient := pkg.NewMysqlClient(node.Config.Mysql)
		node.mysql = mysqlClient
	}

	// 初始化 Redis 客户端
	if node.Config.Redis.Addr != "" {
		rdbClient := pkg.NewRedisClient(&node.Config.Redis)
		node.rdb = rdbClient
	}

	// 初始化 Nsq 客户端
	if node.Config.Nsq.Producer != "" {
		nsqClient := pkg.NewNsqClient(node.Config.Nsq)
		node.nsq = nsqClient
	}

	if node.Config.Rest.Port != 0 {
		node.rest = rest.NewRest(&rest.Proxy{
			Options: &node.Config.Rest,
			Mysql:   node.mysql,
		})
	}

	return node
}

func (n *Node) Name() string {
	return "node"
}

func (n *Node) Init() {
	if n.mysql != nil {
		n.mysql.Init()
	}
	if n.rdb != nil {
		n.rdb.Init()
	}
	if n.nsq != nil {
		n.nsq.Init()
	}

	if n.rest != nil {
		n.rest.Init()
	}

	// 初始化注册中心
	n.registry, _ = consul.NewRegistry(&n.Config.Consul)

}

func (n *Node) Start() error {
	if n.mysql != nil {
		n.mysql.Start()
	}

	if n.rdb != nil {
		n.rdb.Start()
	}

	if n.nsq != nil {
		n.nsq.Start()
	}

	if n.rest != nil {
		n.rest.Start()
	}

	// 启动 gRPC 服务并注册到注册中心
	go func() {
		n.transporter.Start()
	}()
	n.registry.Register(n.NodeId, n.NodeName, n.transporter.GetExposeAddr())
	return nil
}

func (n *Node) Close() {
	if n.mysql != nil {
		n.mysql.Close()
	}
	if n.rdb != nil {
		n.rdb.Close()
	}
	if n.nsq != nil {
		n.nsq.Close()
	}

	if n.rest != nil {
		n.rest.Close()
	}

	n.transporter.Stop()
	n.registry.Close()
}
func (n *Node) Destroy() {
	if n.mysql != nil {
		n.mysql.Destroy()
	}
	if n.rdb != nil {
		n.rdb.Destroy()
	}
	if n.nsq != nil {
		n.nsq.Destroy()
	}

	if n.rest != nil {
		n.rest.Destroy()
	}

	lib.Log.Infof("Node %s is destroyed", n.Config.Id)
}

// / UseEvents 注入事件总线
func (n *Node) UseEvents(events *events.EventBus) {
	n.events = events
}

// 添加grpc服务
func (n *Node) AddServiceProvider(desc *grpc.ServiceDesc, provider any) {
	n.transporter.AddServiceProvider(desc, provider)
}

func (n *Node) GetServiceListen() string {
	return n.transporter.GetListenAddr()
}

// 获取服务地址
func (n *Node) GetServiceAddr() string {
	return n.transporter.GetExposeAddr()
}

// 获取grpc客户端连接
func (n *Node) ServiceClient() (*grpc.ClientConn, error) {
	consulTarget, err := n.registry.GetServiceTarget(n.NodeName)
	if err != nil {
		return nil, err
	}
	return n.transporter.NewClient(consulTarget)
}

// GetMysql 获取 MySQL 客户端实例
func (n *Node) GetMysql() (*pkg.MysqlClient, error) {
	if n.mysql == nil {
		return nil, fmt.Errorf("mysql client is not initialized")
	}
	return n.mysql, nil
}

// GetRdb 获取 Redis 客户端实例
func (n *Node) GetRdb() (*pkg.RedisClient, error) {
	if n.rdb == nil {
		return nil, fmt.Errorf("redis client is not initialized")
	}
	return n.rdb, nil
}

// GetNsq 获取 Nsq 客户端实例
func (n *Node) GetNsq() (*pkg.NsqClient, error) {
	if n.nsq == nil {
		return nil, fmt.Errorf("nsq client is not initialized")
	}
	return n.nsq, nil
}

func (n *Node) GetRest() (*rest.Rest, error) {
	if n.rest == nil {
		return nil, fmt.Errorf("rest is not initialized")
	}
	return n.rest, nil
}

// Proxy 获取节点的代理对象
func (n *Node) Proxy() *Proxy {
	return &Proxy{
		NodeId:   n.NodeId,
		NodeName: n.NodeName,
		Node:     n,
	}
}
