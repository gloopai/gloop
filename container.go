package gloop

import "github.com/gloopai/gloop/component"

type Container struct {
	components []component.Component
}

// NewContainer 创建一个容器
func NewContainer() *Container {
	return &Container{}
}

// Add 添加组件
func (c *Container) Add(components ...component.Component) {
	c.components = append(c.components, components...)
}

// Serve 启动容器
func (c *Container) Serve() {
	c.doInitComponents()

	c.doStartComponents()
}

// 初始化所有组件
func (c *Container) doInitComponents() {
	for _, comp := range c.components {
		comp.Init()
	}
}

// 启动所有组件
func (c *Container) doStartComponents() {
	for _, comp := range c.components {
		comp.Start()
	}
}
