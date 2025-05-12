package node

import (
	"fmt"

	"github.com/gloopai/gloop/modules"
)

type Node struct {
	modules.Base
	Config *NodeConfig
}

type NodeConfig struct {
	NodeID  string `json:"node_id"`
	Address string `json:"address"`
}

func NewNode(config *NodeConfig) *Node {
	return &Node{
		Config: config,
	}
}

func (n *Node) Name() string {
	return "node"
}

func (n *Node) Init() {
}

func (n *Node) Start() error {
	return nil
}
func (n *Node) printInfo() {
	infos := make([]string, 0, 7)
	infos = append(infos, fmt.Sprintf("NodeID: %s", n.Config.NodeID))
	// infos = append(infos, fmt.Sprintf("Name: %s", s.Name()))
	infos = append(infos, fmt.Sprintf("Address: %d", n.Config.NodeID))

	modules.PrintBoxInfo(n.Name(), infos...)
}
