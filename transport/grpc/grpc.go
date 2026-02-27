package grpc

import (
	"log"
	"net"
	"sync"

	"golang.org/x/sync/singleflight"

	"github.com/gloopai/gloop/lib"
	gnet "github.com/gloopai/gloop/lib/net"
	_ "github.com/mbobakov/grpc-consul-resolver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

type Transporter struct {
	listenAddr  string
	exposeAddr  string
	server      *grpc.Server
	sfg         singleflight.Group
	connections sync.Map
}

func NewTransporter(opts *Options) (*Transporter, error) {
	listenAddr, exposeAddr, err := gnet.ParseAddr(opts.Addr)
	if err != nil {
		return nil, err
	}

	// 创建 gRPC 服务器
	s := grpc.NewServer(
	// grpc.UnaryInterceptor(grpcserverlib.NewRateLimiter(1).UnaryInterceptor),
	)
	healthpb.RegisterHealthServer(s, health.NewServer())
	return &Transporter{
		listenAddr: listenAddr,
		exposeAddr: exposeAddr,
		server:     s,
	}, nil
}

// 启动 grpc 服务
func (t *Transporter) Start() error {
	addr, err := net.ResolveTCPAddr("tcp", t.listenAddr)
	if err != nil {
		return err
	}
	lis, err := net.Listen(addr.Network(), addr.String())
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	// lib.Log.Infof("gRPC server is listening at %v export: %v", addr.String(), t.exposeAddr)
	return t.server.Serve(lis)
}

// 关闭 grpc 服务
func (t *Transporter) Stop() {
	t.server.GracefulStop()
	// 关闭所有客户端连接
	t.connections.Range(func(key, value any) bool {
		if conn, ok := value.(*grpc.ClientConn); ok {
			err := conn.Close()
			if err != nil {
				lib.Log.Errorf("Failed to close gRPC client connection for target %s: %v", key, err)
			}
		}
		return true
	})
	lib.Log.Infof("gRPC server is stopped")
}

func (t *Transporter) GetListenAddr() string {
	return t.listenAddr
}

func (t *Transporter) GetExposeAddr() string {
	return t.exposeAddr
}

// 添加服务
func (t *Transporter) AddServiceProvider(name string, desc *grpc.ServiceDesc, provider any) {
	t.server.RegisterService(desc, provider)
}

// NewClient 新建gRPC客户端
func (t *Transporter) NewClient(target string) (*grpc.ClientConn, error) {
	if c, ok := t.connections.Load(target); ok {
		return c.(*grpc.ClientConn), nil
	}
	c, err, _ := t.sfg.Do(target, func() (any, error) {
		defer func() {
			if r := recover(); r != nil {
				lib.Log.Errorf("Recovered from panic in NewClient for target %s: %v", target, r)
			}
		}()
		cc, err := grpc.NewClient(target,
			grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy": "round_robin"}`),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err != nil {
			return nil, err
		}
		t.connections.Store(target, cc)
		return cc, nil
	})
	if err != nil {
		return nil, err
	}
	return c.(*grpc.ClientConn), nil
}
