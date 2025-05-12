package modules

type Component interface {
	// Name 组件名称
	Name() string
	// Init 初始化组件
	Init()
	// Start 启动组件
	Start() error
	// Close 关闭组件
	Close()
	// Destroy 销毁组件
	Destroy()
	// 写入上下文
	SetContext(ctx *ComponentContext)
	// 读取上下文
	GetContext() *ComponentContext
}

type Base struct {
	ctx *ComponentContext
}

// Name 组件名称
func (b *Base) Name() string { return "base" }

// Init 初始化组件
func (b *Base) Init() {}

// Start 启动组件
func (b *Base) Start() error { return nil }

// Close 关闭组件
func (b *Base) Close() {}

// Destroy 销毁组件
func (b *Base) Destroy() {}

// SetContext 写入上下文
func (b *Base) SetContext(ctx *ComponentContext) {
	b.ctx = ctx
}

// GetContext 读取上下文
func (b *Base) GetContext() *ComponentContext {
	return b.ctx
}
