package grpc

import (
	"context"
	"net"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"

	"github.com/gloopai/gloop/lib"
	gnet "github.com/gloopai/gloop/lib/net"
	_ "github.com/mbobakov/grpc-consul-resolver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
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
		return err
	}
	// lib.Log.Infof("gRPC server is listening at %v export: %v", addr.String(), t.ExposeAddr)
	return t.server.Serve(lis)
}

// 关闭 grpc 服务
func (t *Transporter) Stop() {
	if t.server != nil {
		t.server.GracefulStop()
	}
	// 关闭所有客户端连接并清理映射
	t.connections.Range(func(key, value any) bool {
		if conn, ok := value.(*grpc.ClientConn); ok {
			conn.Close()
		}
		t.connections.Delete(key)
		return true
	})
	lib.Log.Infof("gRPC server is stopped")
}

// 添加服务
func (t *Transporter) AddServiceProvider(name string, desc *grpc.ServiceDesc, provider any) {
	t.server.RegisterService(desc, provider)
}

// NewClient 新建gRPC客户端
func (t *Transporter) NewClient(target string) (*grpc.ClientConn, error) {
	// 快速路径：重用现有健康连接
	if v, ok := t.connections.Load(target); ok {
		if cc, ok2 := v.(*grpc.ClientConn); ok2 {
			st := cc.GetState()
			if st == connectivity.Ready || st == connectivity.Idle || st == connectivity.Connecting {
				return cc, nil
			}
			// 如果是 Shutdown 或 TransientFailure，则继续创建新连接
		}
	}

	// 使用 singleflight 避免并发拨号
	c, err, _ := t.sfg.Do(target, func() (any, error) {
		// 如果在等待期间另一个 goroutine 已创建连接，则重用它
		if v, ok := t.connections.Load(target); ok {
			if cc, ok2 := v.(*grpc.ClientConn); ok2 {
				st := cc.GetState()
				if st == connectivity.Ready || st == connectivity.Idle || st == connectivity.Connecting {
					return cc, nil
				}
			}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		cc, err := grpc.NewClient(target,
			grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy": "round_robin"}`),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err != nil {
			return nil, err
		}

		// 等待连接变为 READY，或直到上下文超时
		for {
			st := cc.GetState()
			if st == connectivity.Ready {
				t.connections.Store(target, cc)
				return cc, nil
			}
			// 等待状态变化或上下文超时/取消
			if ok := cc.WaitForStateChange(ctx, st); !ok {
				// context expired/canceled
				if cc.GetState() == connectivity.Ready {
					t.connections.Store(target, cc)
					return cc, nil
				}
				cc.Close()
				return nil, ctx.Err()
			}
			// 循环并检查新状态
		}
	})
	if err != nil {
		return nil, err
	}
	return c.(*grpc.ClientConn), nil
}
