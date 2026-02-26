package modules

import (
	"github.com/gloopai/gloop/lib"
	"github.com/gloopai/gloop/registry"
)

// Node 组件
type Node struct {
	Base
	Config *NodeOptions
}

type NodeOptions struct {
	// 节点 id，全局必须唯一
	Id string
	// 节点名称，便于识别
	Name string
	// 节点grpc 地址，格式为 "ip:port"，如果没设置就随机端口
	Addr string
	// 注册中心配置
	Consul registry.RegistryOptions
}

func NewNode(config *NodeOptions) *Node {
	node := &Node{
		Config: config,
	}
	return node
}

func (n *Node) Close() {
	lib.Log.Infof("Node %s is closing", n.Config.Id)
}
func (n *Node) Destroy() {
	lib.Log.Infof("Node %s is destroyed", n.Config.Id)
}
