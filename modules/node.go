package modules

import (
	"github.com/gloopai/gloop/lib"
	"github.com/gloopai/gloop/registry/consul"
	ggrpc "github.com/gloopai/gloop/transport/grpc"
	"google.golang.org/grpc"
)

// Node 组件
type Node struct {
	Base
	NodeId      string
	NodeName    string
	Config      *NodeOptions
	transporter *ggrpc.Transporter
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
}

func NewNode(config *NodeOptions) *Node {
	node := &Node{
		Config: config,
	}
	node.NodeId = lib.Conf.GetString(config.Id, lib.Generate.Guid())
	node.NodeName = lib.Conf.GetString(config.Name, "node")
	return node
}

func (n *Node) Init() {
	// 初始化 transporter
	transporter, err := ggrpc.NewTransporter(&ggrpc.Options{
		Addr: n.Config.Addr,
	})
	if err != nil {
		lib.Log.Fatalf("failed to create gRPC transport: %v", err)
	}
	n.transporter = transporter

}

func (n *Node) Start() error {
	go func() {
		n.transporter.Start()
	}()
	return nil
}

func (n *Node) Close() {
	n.transporter.Stop()
}
func (n *Node) Destroy() {
	lib.Log.Infof("Node %s is destroyed", n.Config.Id)
}

// 添加grpc服务
func (n *Node) AddServiceProvider(name string, desc *grpc.ServiceDesc, provider any) {
	n.transporter.AddServiceProvider(name, desc, provider)
}
