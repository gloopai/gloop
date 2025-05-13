package modules

import (
	"fmt"

	"github.com/gloopai/gloop/lib"
	"github.com/gloopai/gloop/modules/node"
)

// Node 组件
type Node struct {
	Base
	Config NodeOptions
	Client *node.Client
}

type NodeOptions struct {
	NodeID  string
	Address string
	Gateway string
}

func NewNode() (*Node, error) {
	config := NodeOptions{}

	node := &Node{
		Config: config,
	}

	err := node.loadOptions()
	if err != nil {
		return nil, err
	}
	return node, nil
}

/* 获取节点配置 */
func (n *Node) loadOptions() error {
	err := lib.Conf.LoadTOML("node.toml", &n.Config)
	if err != nil {
		return fmt.Errorf("load node config failed: %w", err)
	}
	if n.Config.NodeID == "" {
		return fmt.Errorf("NodeID is empty")
	}
	if n.Config.Address == "" {
		return fmt.Errorf("address is empty")
	}

	return nil
}

func (n *Node) Name() string {
	return "node"
}

func (n *Node) Init() {
	n.printInfo()
}

func (n *Node) Start() error {
	n.Client = node.NewClient(node.ClientConfig{
		NodeID:  n.Config.NodeID,
		Address: n.Config.Address,
		Gateway: n.Config.Gateway,
	})
	n.Client.Start()

	return nil
}
func (n *Node) printInfo() {
	infos := make([]string, 0, 7)
	infos = append(infos, fmt.Sprintf("NodeID: %s", n.Config.NodeID))
	// infos = append(infos, fmt.Sprintf("Name: %s", s.Name()))
	infos = append(infos, fmt.Sprintf("Address: %s", n.Config.Address))

	PrintBoxInfo(n.Name(), infos...)
}
