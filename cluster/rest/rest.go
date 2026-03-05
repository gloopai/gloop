package rest

import (
	"fmt"
	"os"

	"github.com/gloopai/gloop/events"
	"github.com/gloopai/gloop/lib"
	"github.com/gloopai/gloop/modules/pkg"
	"github.com/gloopai/gloop/modules/rest"
	"github.com/gloopai/gloop/registry/consul"
	ggrpc "github.com/gloopai/gloop/transport/grpc"
	"google.golang.org/grpc"
)

// Gate 组件
type Node struct {
	NodeId      string
	NodeName    string
	Config      *RestOptions
	events      *events.EventBus
	transporter *ggrpc.Transporter
	registry    *consul.Registry
	rdb         *pkg.RedisClient
	mysql       *pkg.MysqlClient
	nsq         *pkg.NsqClient
	rest        *rest.Rest
}

type RestOptions struct {
	// 节点 id，全局必须唯一
	Id string
	// 节点名称，便于识别
	Name string
	// Rest 节点绑定的ip:port，如果没设置就不启动 Rest 服务
	Port int
	// 节点grpc 地址，格式为 "ip:port"，如果没设置就随机端口
	Grpc string
	// 注册中心配置
	Consul consul.Options
	// Redis 配置
	Redis pkg.RedisClientOptions
	// MySQL 配置
	Mysql pkg.MysqlClientOptions
	// Nsq 配置
	Nsq pkg.NsqClientOptions
	// Rest 配置
	Auth rest.SiteAuthOption
}

func NewNode(config *RestOptions) *Node {
	node := &Node{
		Config: config,
	}
	node.NodeId = lib.Conf.GetString(config.Id, lib.Generate.Guid())
	node.NodeName = lib.Conf.GetString(config.Name, "rest")

	if node.Config.Mysql.DSN == "" {
		lib.Log.Warn("[rest] MySQL DSN is not provided, MySQL client will not be initialized")
		os.Exit(0)
	}
	mysqlClient := pkg.NewMysqlClient(node.Config.Mysql)
	node.mysql = mysqlClient

	if node.Config.Redis.Addr == "" {
		lib.Log.Warn("[rest] Redis address is not provided, Redis client will not be initialized")
		os.Exit(0)
	}
	// 初始化 Redis 客户端
	rdbClient := pkg.NewRedisClient(&node.Config.Redis)
	node.rdb = rdbClient

	if node.Config.Nsq.Producer == "" || node.Config.Nsq.Subscribe == "" {
		lib.Log.Warn("[rest] Nsq producer address or subscribe address is not provided, Nsq client will not be initialized")
		os.Exit(0)
	}
	// 初始化 Nsq 客户端
	nsqClient := pkg.NewNsqClient(node.Config.Nsq)
	node.nsq = nsqClient

	if node.Config.Port != 0 {
		node.rest = rest.NewRest(&rest.Proxy{
			Options: &rest.RestOptions{
				Port: node.Config.Port,
				Auth: node.Config.Auth,
			},
			Mysql: node.mysql,
		})
	}

	// 初始化 GRpc transporter
	transporter, err := ggrpc.NewTransporter(&ggrpc.Options{
		Addr: node.Config.Grpc,
	})
	if err != nil {
		lib.Log.Fatalf("failed to create gRPC transport: %v", err)
	}
	node.transporter = transporter

	return node
}

func (g *Node) Name() string {
	return "gate"
}

func (g *Node) Init() {
	if g.mysql != nil {
		g.mysql.Init()
	}
	if g.rdb != nil {
		g.rdb.Init()
	}
	if g.nsq != nil {
		g.nsq.Init()
	}

	if g.rest != nil {
		g.rest.Init()
	}

	// 初始化注册中心
	g.registry, _ = consul.NewRegistry(&g.Config.Consul)

}

func (g *Node) Start() error {
	if g.mysql != nil {
		g.mysql.Start()
	}

	if g.rdb != nil {
		g.rdb.Start()
	}

	if g.nsq != nil {
		g.nsq.Start()
	}

	if g.rest != nil {
		g.rest.Start()
	}

	// 启动 gRPC 服务并注册到注册中心
	go func() {
		g.transporter.Start()
	}()
	g.registry.Register(g.NodeId, g.NodeName, g.transporter.GetExposeAddr())
	return nil
}

func (g *Node) Close() {
	if g.mysql != nil {
		g.mysql.Close()
	}
	if g.rdb != nil {
		g.rdb.Close()
	}
	if g.nsq != nil {
		g.nsq.Close()
	}

	if g.rest != nil {
		g.rest.Close()
	}

	g.transporter.Stop()
	g.registry.Close()
}
func (g *Node) Destroy() {
	if g.mysql != nil {
		g.mysql.Destroy()
	}
	if g.rdb != nil {
		g.rdb.Destroy()
	}
	if g.nsq != nil {
		g.nsq.Destroy()
	}

	if g.rest != nil {
		g.rest.Destroy()
	}

	lib.Log.Infof("Node %s is destroyed", g.Config.Id)
}

// / UseEvents 注入事件总线
func (g *Node) UseEvents(events *events.EventBus) {
	g.events = events
}

// 添加grpc服务
func (g *Node) AddServiceProvider(desc *grpc.ServiceDesc, provider any) {
	g.transporter.AddServiceProvider(desc, provider)
}

func (g *Node) GetServiceListen() string {
	return g.transporter.GetListenAddr()
}

// 获取服务地址
func (g *Node) GetServiceAddr() string {
	return g.transporter.GetExposeAddr()
}

// 获取grpc客户端连接
func (g *Node) ServiceClient() (*grpc.ClientConn, error) {
	consulTarget, err := g.registry.GetServiceTarget(g.NodeName)
	if err != nil {
		return nil, err
	}
	return g.transporter.NewClient(consulTarget)
}

// GetMysql 获取 MySQL 客户端实例
func (g *Node) GetMysql() (*pkg.MysqlClient, error) {
	if g.mysql == nil {
		return nil, fmt.Errorf("mysql client is not initialized")
	}
	return g.mysql, nil
}

// GetRdb 获取 Redis 客户端实例
func (g *Node) GetRdb() (*pkg.RedisClient, error) {
	if g.rdb == nil {
		return nil, fmt.Errorf("redis client is not initialized")
	}
	return g.rdb, nil
}

// GetNsq 获取 Nsq 客户端实例
func (g *Node) GetNsq() (*pkg.NsqClient, error) {
	if g.nsq == nil {
		return nil, fmt.Errorf("nsq client is not initialized")
	}
	return g.nsq, nil
}

func (g *Node) GetRest() (*rest.Rest, error) {
	if g.rest == nil {
		return nil, fmt.Errorf("rest is not initialized")
	}
	return g.rest, nil
}

// Proxy 获取节点的代理对象
func (g *Node) Proxy() *Proxy {
	return &Proxy{
		NodeId:   g.NodeId,
		NodeName: g.NodeName,
		Node:     g,
	}
}
