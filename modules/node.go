package modules

import (
	"fmt"
	"time"

	"github.com/gloopai/gloop/lib"
)

// Client
type Client struct {
	Config *ClientConfig
	ticker *time.Ticker
}

type ClientConfig struct {
	CenterAddress string `json:"center_address"`
	NodeID        string `json:"node_id"`
	Address       string `json:"address"`
}

func NewClient(conf ClientConfig) *Client {
	return &Client{
		Config: &conf,
	}
}

func (c *Client) Start() {
	c.ticker = time.NewTicker(5 * time.Second) // 每 5 秒触发一次
	go func() {
		for range c.ticker.C {
			c.Heartbeat()
		}
	}()
}

func (c *Client) Stop() {
	if c.ticker != nil {
		c.ticker.Stop()
	}
}

func (c *Client) Register() error {
	return nil
}

func (c *Client) Heartbeat() {
	// 心跳逻辑
	// 例如：向 CenterAddress 发送心跳请求
	println("Heartbeat sent to", c.Config.CenterAddress)
}

// Node 组件
type Node struct {
	Base
	Config NodeOptions
	Client *Client
}

type NodeOptions struct {
	NodeID  string
	Address string
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
	n.Client = NewClient(ClientConfig{})
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
