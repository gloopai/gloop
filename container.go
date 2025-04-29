package gloop

import "github.com/gloogai/gloop/component"

type Container struct {
	components []component.Component
}

// NewContainer 创建一个容器
func NewContainer() *Container {
	return &Container{}
}
