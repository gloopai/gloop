package gloop

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"

	"github.com/gloopai/gloop/cluster/gate"
	"github.com/gloopai/gloop/component"
	"github.com/gloopai/gloop/events"
	"github.com/gloopai/gloop/lib"
	"github.com/gloopai/gloop/modules"
)

type ContainerProxy struct {
	Gate *gate.Node
}

type Container struct {
	Config     *ContainerConfig
	components []component.Component
	events     *events.EventBus
	gate       *gate.Node
}

type ContainerConfig struct {
	LogLevel lib.LogLevel
	Debug    bool
	Gate     gate.GateOptions
}

// NewContainer 创建一个容器
func NewContainer() *Container {
	config, err := loadOptions()
	if err != nil {
		log.Fatalf("[NewContainer] Failed to load container configuration: %v", err)
	}
	lib.Log.SetLogLevel(config.LogLevel)
	lib.Log.SetDebugEnabled(config.Debug)
	lib.Log.InitLogger(config.LogLevel, nil)

	c := &Container{
		Config: config,
		events: events.NewEventBus(),
	}
	// 初始化 Gate 组件
	c.gate = gate.NewNode(&config.Gate)
	c.gate.UseEvents(c.events) // 注入事件总线

	return c
}

func loadOptions() (*ContainerConfig, error) {
	var options *ContainerConfig
	err := lib.Conf.LoadTOML("container.toml", &options)
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %v", err)
	}

	return options, nil
}

// Add 添加组件
func (c *Container) Add(components ...component.Component) {
	c.components = append(c.components, components...)
}

// Serve 启动容器
func (c *Container) Serve() {
	c.doPrintFrameworkInfo()
	c.doInitComponents()
	c.doRegisterComponents()
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

// 初始化 container 默认组件
func (c *Container) initDefaultComponents() {
	if c.gate != nil {
		c.gate.Init()
	}
}

// 初始化所有组件
func (c *Container) doInitComponents() {
	c.initDefaultComponents()
	for _, comp := range c.components {
		defer func() {
			if r := recover(); r != nil {
				lib.Log.Errorf("Recovered from panic in component %s: %v Init\nStack trace:\n%s", comp.Name(), r, string(debug.Stack()))
			}
		}()
		comp.Init()
	}
	lib.Log.Info("🟢 Components INIT Complete!!")
}

// 注册 grpc 服务
func (c *Container) doRegisterComponents() {
	for _, comp := range c.components {
		defer func() {
			if r := recover(); r != nil {
				lib.Log.Errorf("Recovered from panic in component %s: %v Register\nStack trace:\n%s", comp.Name(), r, string(debug.Stack()))
			}
		}()
		comp.Register()
	}
	lib.Log.Info("🟢 Components Service REGISTER Complete!!")
}

// 启动 container 默认组件
func (c *Container) startDefaultComponents() {
	if c.gate != nil {
		c.gate.Start()
	}
}

// 启动所有组件
func (c *Container) doStartComponents() {
	c.startDefaultComponents()
	for _, comp := range c.components {
		defer func() {
			if r := recover(); r != nil {
				lib.Log.Errorf("Recovered from panic in component %s: %v Start\nStack trace:\n%s", comp.Name(), r, string(debug.Stack()))
			}
		}()
		if err := comp.Start(); err != nil {
			lib.Log.Errorf("Failed to start component %s: %v\nStack trace:\n%s", comp.Name(), err, string(debug.Stack()))
		}
	}

	lib.Log.Info("🟢 Components START Complete!!")
}

// 销毁 container 默认组件
func (c *Container) destroyDefaultComponents() {
	if c.gate != nil {
		c.gate.Close()
		c.gate.Destroy()
	}
}

// 销毁所有组件
func (c *Container) doDestroyComponents() {
	c.destroyDefaultComponents()

	for _, comp := range c.components {
		defer func() {
			if r := recover(); r != nil {
				lib.Log.Errorf("Recovered from panic in component %s: %v Destroy\nStack trace:\n%s", comp.Name(), r, string(debug.Stack()))
			}
		}()
		comp.Destroy()
	}
	lib.Log.Info("🟢 Components DESTROY Complete!!")
}

// 打印框架信息
func (c *Container) doPrintFrameworkInfo() {
	modules.PrintFrameworkInfo()

	infos := make([]string, 0, 7)
	if c.gate != nil {
		infos = append(infos, fmt.Sprintf("Node ID: %s", c.gate.NodeId))
		infos = append(infos, fmt.Sprintf("Node Name: %s", c.gate.NodeName))
		if c.gate != nil {
			infos = append(infos, fmt.Sprintf("gRPC Listen: %s", c.gate.GetServiceListen()))
			infos = append(infos, fmt.Sprintf("gRPC Expose: %s", c.gate.GetServiceAddr()))
		}
	}
	infos = append(infos, fmt.Sprintf("Debug: %v", c.Config.Debug))
	infos = append(infos, fmt.Sprintf("LogLevel: %v", c.Config.LogLevel))
	if c.Config.Gate.Mysql.DSN != "" {
		infos = append(infos, fmt.Sprintf("Mysql: %v", true))
	}
	if c.Config.Gate.Redis.Addr != "" {
		infos = append(infos, fmt.Sprintf("Redis: %s", c.Config.Gate.Redis.Addr))
	}

	if c.Config.Gate.Rest.Port != 0 {
		infos = append(infos, fmt.Sprintf("Rest Port: %d", c.Config.Gate.Rest.Port))
	}
	modules.PrintBoxInfo("Container", infos...)
}

// Proxy 获取容器的代理对象
func (c *Container) Proxy() *ContainerProxy {
	return &ContainerProxy{
		Gate: c.gate,
	}
}
