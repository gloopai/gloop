package gloop

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gloopai/gloop/events"
	"github.com/gloopai/gloop/lib"
	"github.com/gloopai/gloop/modules"
	"github.com/gloopai/gloop/modules/site"
)

type ContainerProxy struct {
	Site *site.Site
}

type Container struct {
	Config     *ContainerConfig
	components []modules.Component
	Node       *modules.Node
	Database   *modules.MysqlClient
	Rdb        *modules.Rdb
	EventBus   *events.EventBus
	Site       *site.Site
}

type ContainerConfig struct {
	LogLevel lib.LogLevel
	Debug    bool
	Db       modules.MysqlClientOptions
	Redis    modules.RdbOptions
	Site     site.SiteOptions
	Node     modules.NodeOptions
}

// NewContainer 创建一个容器
func NewContainer() *Container {
	config, err := loadOptions()
	if err != nil {
		log.Fatalf("[NewContainer] Failed to load container configuration: %v", err)
	}
	lib.Log.SetLogLevel(config.LogLevel)
	lib.Log.SetDebugEnabled(config.Debug)
	// Initialize logger with default formatter (includes timestamps)
	lib.Log.InitLogger(config.LogLevel, nil)

	c := &Container{
		Config:   config,
		EventBus: events.NewEventBus(),
	}

	// 数据库链接
	if config.Db.DSN != "" {
		dbService := modules.NewMysqlClient(config.Db)
		c.Database = dbService
	}

	// 初始化 Redis 组件
	if config.Redis.Addr != "" {
		rdb := modules.NewRdb(&config.Redis)
		c.Rdb = rdb
	}

	// 初始化 Site 组件
	if config.Site.Port != 0 {
		c.Site = site.NewSite(config.Site)
		c.Site.SetEnv(&modules.ComponentEnv{
			DB:     c.Database,
			Rdb:    c.Rdb,
			Events: c.EventBus,
		})
	}

	// 初始化 Node 组件
	c.Node = modules.NewNode(&config.Node)
	c.Node.SetEnv(&modules.ComponentEnv{
		DB:     c.Database,
		Rdb:    c.Rdb,
		Events: c.EventBus,
	})
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
func (c *Container) Add(components ...modules.Component) {
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
	if c.Database != nil {
		c.Database.Init()
	}
	if c.Rdb != nil {
		c.Rdb.Init()
	}
	if c.Site != nil {
		c.Site.Init()
	}

	if c.Node != nil {
		c.Node.Init()
	}
}

// 初始化所有组件
func (c *Container) doInitComponents() {
	c.initDefaultComponents()
	for _, comp := range c.components {
		defer func() {
			if r := recover(); r != nil {
				lib.Log.Errorf("Recovered from panic in component %s: %v", comp.Name(), r)
			}
		}()
		comp.SetEnv(&modules.ComponentEnv{
			Node:   c.Node,
			DB:     c.Database,
			Rdb:    c.Rdb,
			Events: c.EventBus,
		})
		comp.Init()
	}
	lib.Log.Info("🟢 Components INIT Complete!!")
}

// 注册 grpc 服务
func (c *Container) doRegisterComponents() {
	for _, comp := range c.components {
		defer func() {
			if r := recover(); r != nil {
				lib.Log.Errorf("Recovered from panic in component %s: %v", comp.Name(), r)
			}
		}()
		comp.Register()
	}
	lib.Log.Info("🟢 Components Service REGISTER Complete!!")
}

// 启动 container 默认组件
func (c *Container) startDefaultComponents() {
	if c.Database != nil {
		c.Database.Start()
	}
	if c.Rdb != nil {
		c.Rdb.Start()
	}

	if c.Site != nil {
		c.Site.Start()
	}
	if c.Node != nil {
		c.Node.Start()
	}
}

// 启动所有组件
func (c *Container) doStartComponents() {
	c.startDefaultComponents()
	for _, comp := range c.components {
		defer func() {
			if r := recover(); r != nil {
				lib.Log.Errorf("Recovered from panic in component %s: %v", comp.Name(), r)
			}
		}()
		if err := comp.Start(); err != nil {
			lib.Log.Errorf("Failed to start component %s: %v", comp.Name(), err)
		}
	}

	lib.Log.Info("🟢 Components START Complete!!")
}

// 销毁 container 默认组件
func (c *Container) destroyDefaultComponents() {
	if c.Site != nil {
		c.Site.Close()
		c.Site.Destroy()
	}
	if c.Database != nil {
		c.Database.Close()
	}
	if c.Rdb != nil {
		c.Rdb.Close()
	}
	if c.Node != nil {
		c.Node.Close()
		c.Node.Destroy()
	}
}

// 销毁所有组件
func (c *Container) doDestroyComponents() {
	c.destroyDefaultComponents()

	for _, comp := range c.components {
		defer func() {
			if r := recover(); r != nil {
				lib.Log.Errorf("Recovered from panic in component %s: %v", comp.Name(), r)
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
	if c.Node != nil {
		infos = append(infos, fmt.Sprintf("Node ID: %s", c.Node.NodeId))
		infos = append(infos, fmt.Sprintf("Node Name: %s", c.Node.NodeName))
		if c.Node != nil {
			infos = append(infos, fmt.Sprintf("gRPC Listen: %s", c.Node.GetServiceListen()))
			infos = append(infos, fmt.Sprintf("gRPC Expose: %s", c.Node.GetServiceAddr()))
		}
	}
	infos = append(infos, fmt.Sprintf("Debug: %v", c.Config.Debug))
	infos = append(infos, fmt.Sprintf("LogLevel: %v", c.Config.LogLevel))
	if c.Database != nil {
		infos = append(infos, fmt.Sprintf("Database: %s", "mysql"))
	}
	if c.Rdb != nil {
		infos = append(infos, fmt.Sprintf("Redis: %s", c.Config.Redis.Addr))
	}

	if c.Config.Site.Port != 0 {
		infos = append(infos, fmt.Sprintf("Site Port: %d", c.Config.Site.Port))
	}
	modules.PrintBoxInfo("Container", infos...)
}

// Proxy 获取容器的代理对象
func (c *Container) Proxy() *ContainerProxy {
	return &ContainerProxy{
		Site: c.Site,
	}
}
