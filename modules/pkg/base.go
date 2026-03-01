package pkg

type Base struct {
	Env *interface{}
}

// Name 组件名称
func (b *Base) Name() string { return "pkg base" }

// Init 初始化组件
func (b *Base) Init() {}

// 注册服务
func (b *Base) Register() {}

// Start 启动组件
func (b *Base) Start() error { return nil }

// Close 关闭组件
func (b *Base) Close() {}

// Destroy 销毁组件
func (b *Base) Destroy() {}

// SetEnv 写入上下文
func (b *Base) SetEnv(env *interface{}) {
	b.Env = env
}

// GetEnv 读取上下文
func (b *Base) GetEnv() *interface{} {
	return b.Env
}
