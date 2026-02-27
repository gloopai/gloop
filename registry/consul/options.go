package consul

import (
	"context"

	"github.com/hashicorp/consul/api"
)

type Options struct {
	// 客户端连接地址
	// 内建客户端配置，默认为127.0.0.1:8500
	Addr string

	// 外部客户端
	// 外部客户端配置，存在外部客户端时，优先使用外部客户端，默认为nil
	Client *api.Client

	// 上下文
	// 默认为context.Background
	Ctx context.Context

	// 是否启用健康检查
	// 默认为true
	EnableHealthCheck bool

	// 健康检查时间间隔（秒），仅在启用健康检查后生效
	// 默认10秒
	HealthCheckInterval int

	// 健康检查超时时间（秒），仅在启用健康检查后生效
	// 默认5秒
	HealthCheckTimeout int

	// 是否启用心跳检查
	// 默认为true
	EnableHeartbeatCheck bool

	// 心跳检查时间间隔（秒），仅在启用心跳检查后生效
	// 默认10秒
	HeartbeatCheckInterval int

	// 健康检测失败后自动注销服务时间（秒）
	// 默认30秒
	DeregisterCriticalServiceAfter int
}
