package modules

import (
	"errors"
	"sync"
	"time"

	"github.com/gloopai/gloop/lib"
	"github.com/gloopai/gloop/registry/consul"
	ggrpc "github.com/gloopai/gloop/transport/grpc"
	"google.golang.org/grpc"
)

// Node 组件
type Node struct {
	Base
	NodeId      string
	NodeName    string
	Config      *NodeOptions
	transporter *ggrpc.Transporter
	registry    *consul.Registry
	mu          sync.RWMutex
	initialized bool
	initErr     error
	started     bool
}

type NodeOptions struct {
	// 节点 id，全局必须唯一
	Id string
	// 节点名称，便于识别
	Name string
	// 节点grpc 地址，格式为 "ip:port"，如果没设置就随机端口
	Addr string
	// 注册中心配置
	Consul consul.Options
}

func NewNode(config *NodeOptions) *Node {
	if config == nil {
		config = &NodeOptions{}
	}
	node := &Node{Config: config}
	if config.Id != "" {
		node.NodeId = config.Id
	} else {
		node.NodeId = lib.Generate.Guid()
	}
	if config.Name != "" {
		node.NodeName = config.Name
	} else {
		node.NodeName = "node"
	}
	return node
}

func (n *Node) Init() {
	if n.Config == nil {
		n.initErr = errors.New("node config is nil")
		lib.Log.Errorf("Node.Init: %v", n.initErr)
		return
	}

	// 初始化 transporter
	transporter, err := ggrpc.NewTransporter(&ggrpc.Options{
		Addr: n.Config.Addr,
	})
	if err != nil {
		n.initErr = err
		lib.Log.Errorf("failed to create gRPC transport: %v", err)
		return
	}

	// 初始化注册中心
	registry, err := consul.NewRegistry(&n.Config.Consul)
	if err != nil {
		n.initErr = err
		lib.Log.Errorf("failed to create consul registry: %v", err)
		return
	}

	n.mu.Lock()
	n.transporter = transporter
	n.registry = registry
	n.initialized = true
	n.initErr = nil
	n.mu.Unlock()
}

func (n *Node) Start() error {
	n.mu.RLock()
	if n.initErr != nil {
		err := n.initErr
		n.mu.RUnlock()
		return err
	}
	if !n.initialized || n.transporter == nil || n.registry == nil {
		n.mu.RUnlock()
		return errors.New("node not initialized")
	}
	n.mu.RUnlock()

	n.mu.Lock()
	if n.started {
		n.mu.Unlock()
		return nil
	}
	n.started = true
	n.mu.Unlock()

	// Start transporter in background and capture error
	errCh := make(chan error, 1)
	go func() {
		if err := n.transporter.Start(); err != nil {
			errCh <- err
		}
		close(errCh)
	}()

	// wait until transporter exposes port or returns error
	timeout := time.After(5 * time.Second)
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case err := <-errCh:
			if err != nil {
				return err
			}
			// transporter stopped without setting port
			return errors.New("transporter stopped unexpectedly")
		case <-ticker.C:
			if n.transporter.ExposePort != 0 {
				// ready
				goto REGISTER
			}
		case <-timeout:
			return errors.New("timeout waiting for transporter to be ready")
		}
	}

REGISTER:
	if err := n.registry.Register(n.NodeId, n.NodeName, n.transporter.ExposeAddr, n.transporter.ExposePort); err != nil {
		// attempt to stop transporter if register failed
		n.transporter.Stop()
		return err
	}
	return nil
}

func (n *Node) Close() {
	n.mu.Lock()
	if n.registry != nil {
		if err := n.registry.Close(); err != nil {
			lib.Log.Warnf("failed to deregister service: %v", err)
		}
	}
	if n.transporter != nil {
		n.transporter.Stop()
	}
	n.started = false
	n.mu.Unlock()
}
func (n *Node) Destroy() {
	lib.Log.Infof("Node %s is destroyed", n.Config.Id)
}

// 添加grpc服务
func (n *Node) AddServiceProvider(name string, desc *grpc.ServiceDesc, provider any) {
	n.transporter.AddServiceProvider(name, desc, provider)
}

func (n *Node) GetServiceListen() string {
	return n.transporter.ListenAddr
}

// 获取服务地址
func (n *Node) GetServiceAddr() string {
	return n.transporter.ExposeAddr
}

// 获取服务端口
func (n *Node) GetServicePort() int {
	return n.transporter.ExposePort
}

// 获取grpc客户端连接
func (n *Node) ServiceClient(target string) (*grpc.ClientConn, error) {
	consulTarget, err := n.registry.GetServiceTarget(target)
	if err != nil {
		return nil, err
	}
	return n.transporter.NewClient(consulTarget)
}
