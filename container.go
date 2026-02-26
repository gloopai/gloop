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
	"github.com/gloopai/gloop/registry"
	"github.com/gloopai/gloop/registry/consul"
)

type ContainerProxy struct {
	Site *site.Site
}

type Container struct {
	Config     *ContainerConfig
	components []modules.Component
	Node       *modules.Node
	Database   *modules.DbService
	EventBus   *events.EventBus
	Site       *site.Site
	Registry   *registry.Registry
}

type ContainerConfig struct {
	LogLevel lib.LogLevel
	Debug    bool
	Db       modules.DbOptions
	Site     site.SiteOptions
	Registry consul.ConsulOptions
}

// NewContainer 创建一个容器
func NewContainer() *Container {
	config, err := loadOptions()
	if err != nil {
		log.Fatalf("[NewContainer] Failed to load container configuration: %v", err)
	}
	lib.Log.SetLogLevel(config.LogLevel)
	lib.Log.SetDebugEnabled(config.Debug)

	c := &Container{
		Config:   config,
		EventBus: events.NewEventBus(),
	}

	// 数据库链接
	if config.Db.DSN != "" {
		dbService := modules.NewDb(config.Db)
		dbService.Init()
		c.Database = dbService
	}

	// 初始化 Site 组件
	if config.Site.Port != 0 {
		c.Site = site.NewSite(config.Site)
		c.Site.SetEnv(&modules.ComponentEnv{
			DB:     c.Database,
			Events: c.EventBus,
		})
		c.Site.Init()
		c.Site.Start()
	}

	// 初始化 Registry 组件
	if config.Registry.Addr != "" {
		c.Registry = registry.NewRegistry(config.Registry)
		c.Registry.Init()
		c.Registry.Start()
	}

	// node, err := modules.NewNode()
	// if err != nil {
	// 	lib.Log.Fatal(err)
	// 	os.Exit(0)
	// }
	// c.Node = node

	return c
}

func loadOptions() (*ContainerConfig, error) {
	var options *ContainerConfig
	err := lib.Conf.LoadTOML("container.toml", &options)
	if err != nil {
		return nil, fmt.Errorf("failed to load webhook configuration: %v", err)
	}

	return options, nil
}

func (c *Container) destroy() {
	if c.Site != nil {
		c.Site.Close()
		c.Site.Destory()
	}
	if c.Registry != nil {
		c.Registry.Close()
		c.Registry.Destroy()
	}
	if c.Database != nil {
		c.Database.Close()
	}
}

// Add 添加组件
func (c *Container) Add(components ...modules.Component) {
	c.components = append(c.components, components...)
}

// Serve 启动容器
func (c *Container) Serve() {
	c.doPrintFrameworkInfo()
	// 初始化节点
	// c.Node.Init()
	// c.Node.Start()

	c.doInitComponents()
	c.doRegComponentsService()
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
		comp.SetEnv(&modules.ComponentEnv{
			Node:   c.Node,
			DB:     c.Database,
			Events: c.EventBus,
		})
		comp.Init()
	}
}

// 初始化所有组件
func (c *Container) doRegComponentsService() {
	for _, comp := range c.components {
		comp.RegisterService()
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
	c.destroy()

	for _, comp := range c.components {
		comp.Destroy()
	}
	lib.Log.Info("Container stopped")
}

// 打印框架信息
func (c *Container) doPrintFrameworkInfo() {
	modules.PrintFrameworkInfo()

	infos := make([]string, 0, 7)
	infos = append(infos, fmt.Sprintf("Debug: %v", c.Config.Debug))
	infos = append(infos, fmt.Sprintf("LogLevel: %v", c.Config.LogLevel))
	if c.Database != nil {
		infos = append(infos, fmt.Sprintf("Database: %s", "mysql"))
	}
	if c.Config.Site.Port != 0 {
		infos = append(infos, fmt.Sprintf("Site Port: %d", c.Config.Site.Port))
	}
	if c.Config.Registry.Addr != "" {
		infos = append(infos, fmt.Sprintf("Registry Addr: %s", c.Config.Registry.Addr))
	}
	modules.PrintBoxInfo("Container", infos...)
}

// Proxy 获取容器的代理对象
func (c *Container) Proxy() *ContainerProxy {
	return &ContainerProxy{
		Site: c.Site,
	}
}
