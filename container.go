package gloop

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/gloopai/gloop/lib"
	"github.com/gloopai/gloop/modules"
)

type Container struct {
	components []modules.Component
	Node       *modules.Node
}

type ContainerConfig struct {
	LogLevel lib.LogLevel
	Debug    bool
}

// NewContainer 创建一个容器
func NewContainer(config ContainerConfig) *Container {
	lib.Log.SetLogLevel(config.LogLevel)
	lib.Log.SetDebugEnabled(config.Debug)

	c := &Container{}

	c.Node = modules.NewNode()
	return c
}

// Add 添加组件
func (c *Container) Add(components ...modules.Component) {
	c.components = append(c.components, components...)
}

// Serve 启动容器
func (c *Container) Serve() {
	c.doPrintFrameworkInfo()
	// 初始化节点
	c.Node.Init()
	c.Node.Start()

	c.doInitComponents()
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

}
