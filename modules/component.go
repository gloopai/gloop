package modules

type Component interface {
	// Name 组件名称
	Name() string
	// Init 初始化组件
	Init()
	// 注册服务
	RegisterService()
	// Start 启动组件
	Start() error
	// Close 关闭组件
	Close()
	// Destroy 销毁组件
	Destroy()
	// 写入上下文
	SetEnv(ctx *ComponentEnv)
	// 读取上下文
	GetEnv() *ComponentEnv
}

type Base struct {
	Env *ComponentEnv
}

// Name 组件名称
func (b *Base) Name() string { return "base" }

// Init 初始化组件
func (b *Base) Init() {}

func (b *Base) RegisterService() {}

// Start 启动组件
func (b *Base) Start() error { return nil }

// Close 关闭组件
func (b *Base) Close() {}

// Destroy 销毁组件
func (b *Base) Destroy() {}

// SetEnv 写入上下文
func (b *Base) SetEnv(env *ComponentEnv) {
	b.Env = env
}

// GetEnv 读取上下文
func (b *Base) GetEnv() *ComponentEnv {
	return b.Env
}
