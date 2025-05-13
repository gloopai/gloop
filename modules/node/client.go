package node

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
	Gateway string `json:"gateway"`
	NodeID  string `json:"node_id"`
	Address string `json:"address"`
}

func NewClient(conf ClientConfig) *Client {
	return &Client{
		Config: &conf,
	}
}

func (c *Client) Start() {
	c.Register()

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

func (c *Client) do(url string, data interface{}) ([]byte, error) {
	header := map[string]interface{}{
		"NodeID":        c.Config.NodeID,
		"Authorization": "Bearer:f846b6c62747dc282d569aba2ee6c117",
	}
	url = c.Config.Gateway + url
	return lib.Request.HttpPostJson(url, data, header)
}

func (c *Client) Register() error {
	body, err := c.do("/nodes/register", map[string]interface{}{
		"NodeID":  c.Config.NodeID,
		"Address": c.Config.Address,
	})
	if err != nil {
		lib.Log.Error("Register error:", err)
	}
	fmt.Println("Register response:", string(body))
	return nil
}

func (c *Client) Heartbeat() {
	// 心跳逻辑
	// 例如：向 CenterAddress 发送心跳请求
	println("Heartbeat sent to", c.Config.Gateway)
}
