package grpc

import (
	"log"
	"net"

	"github.com/gloopai/gloop/lib"
	gnet "github.com/gloopai/gloop/lib/net"
	"google.golang.org/grpc"
)

type Transport struct {
	listenAddr string
	exposeAddr string
	server     *grpc.Server
}

func NewTransport(opts *Options) (*Transport, error) {
	listenAddr, exposeAddr, err := gnet.ParseAddr(opts.Addr)
	if err != nil {
		return nil, err
	}

	// 创建 gRPC 服务器
	s := grpc.NewServer(
	// grpc.UnaryInterceptor(grpcserverlib.NewRateLimiter(1).UnaryInterceptor),
	)
	return &Transport{
		listenAddr: listenAddr,
		exposeAddr: exposeAddr,
		server:     s,
	}, nil
}

// 启动 grpc 服务
func (t *Transport) Start() error {
	addr, err := net.ResolveTCPAddr("tcp", t.listenAddr)
	if err != nil {
		return err
	}

	lis, err := net.Listen(addr.Network(), addr.String())
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	lib.Log.Infof("gRPC server is listening at %v", lis.Addr())
	return t.server.Serve(lis)
}

// 关闭 grpc 服务
func (t *Transport) Stop() {
	t.server.GracefulStop()
	lib.Log.Infof("gRPC server is stopped")
}
