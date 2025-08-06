package gloop

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gloopai/gloop/lib"
	"github.com/gloopai/gloop/modules"
)

type Container struct {
	Config     *ContainerConfig
	components []modules.Component
	Node       *modules.Node
}

type ContainerConfig struct {
	LogLevel lib.LogLevel
	Debug    bool
}

// NewContainer 创建一个容器
func NewContainer() *Container {
	config, err := loadWebhookOptions()
	if err != nil {
		log.Fatalf("[NewWebhook] Failed to load webhook configuration: %v", err)
	}
	lib.Log.SetLogLevel(config.LogLevel)
	lib.Log.SetDebugEnabled(config.Debug)

	c := &Container{
		Config: config,
	}

	// node, err := modules.NewNode()
	// if err != nil {
	// 	lib.Log.Fatal(err)
	// 	os.Exit(0)
	// }
	// c.Node = node

	return c
}

func loadWebhookOptions() (*ContainerConfig, error) {
	var options *ContainerConfig
	err := lib.Conf.LoadTOML("container.toml", &options)
	if err != nil {
		return nil, fmt.Errorf("failed to load webhook configuration: %v", err)
	}

	return options, nil
}

// Add 添加组件
func (c *Container) Add(components ...modules.Component) {
	c.components = append(c.components, components...)
}

// Serve 启动容器
func (c *Container) Serve() {
	c.doPrintFrameworkInfo()
	// 初始化节点
	// c.Node.Init()
	// c.Node.Start()

	c.doInitComponents()
	c.doRegComponentsService()
	c.doStartComponents()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-signalChan
		c.doDestroyComponents()
		os.Exit(0)
	}()

	// Keep the program running
	select {}
}

// 初始化所有组件
func (c *Container) doInitComponents() {
	for _, comp := range c.components {
		comp.SetContext(&modules.ComponentContext{
			Node: c.Node,
		})
		comp.Init()
	}
}

// 初始化所有组件
func (c *Container) doRegComponentsService() {
	for _, comp := range c.components {
		comp.SetContext(&modules.ComponentContext{
			Node: c.Node,
		})
		comp.RegisterService()
	}
}

// 启动所有组件
func (c *Container) doStartComponents() {
	for _, comp := range c.components {
		go comp.Start()
	}
}

// 销毁所有组件
func (c *Container) doDestroyComponents() {
	for _, comp := range c.components {
		comp.Destroy()
	}
}

// 打印框架信息
func (c *Container) doPrintFrameworkInfo() {
	modules.PrintFrameworkInfo()

	infos := make([]string, 0, 7)
	infos = append(infos, fmt.Sprintf("Debug: %v", c.Config.Debug))
	infos = append(infos, fmt.Sprintf("LogLevel: %v", c.Config.LogLevel))
	modules.PrintBoxInfo("Container", infos...)
}
