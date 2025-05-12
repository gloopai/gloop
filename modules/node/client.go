package node

import "time"

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
