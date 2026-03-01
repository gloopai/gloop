package pkg

type Base struct {
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
