package grpc

import (
	"log"
	"net"
	"sync"

	"golang.org/x/sync/singleflight"

	"github.com/gloopai/gloop/lib"
	gnet "github.com/gloopai/gloop/lib/net"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

type Transporter struct {
	ListenAddr  string
	ExposeAddr  string
	ExposePort  int
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
		ListenAddr: listenAddr,
		ExposeAddr: exposeAddr,
		server:     s,
	}, nil
}

// 启动 grpc 服务
func (t *Transporter) Start() error {
	addr, err := net.ResolveTCPAddr("tcp", t.ListenAddr)
	if err != nil {
		return err
	}
	t.ExposePort = addr.Port

	lis, err := net.Listen(addr.Network(), addr.String())
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	// lib.Log.Infof("gRPC server is listening at %v export: %v", addr.String(), t.ExposeAddr)
	return t.server.Serve(lis)
}

// 关闭 grpc 服务
func (t *Transporter) Stop() {
	t.server.GracefulStop()
	lib.Log.Infof("gRPC server is stopped")
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
		cc, err := grpc.NewClient(target,
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
