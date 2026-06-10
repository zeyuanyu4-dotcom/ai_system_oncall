// Package grpcserver 集中管理 gRPC 服务端的启动、拦截器和实现。
package grpcserver

import (
	"context"
	"fmt"
	"net"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// Server 封装 gRPC 服务端
type Server struct {
	srv  *grpc.Server
	addr string
	lis  net.Listener
	log  *zap.Logger
}

// NewServer 创建 gRPC 服务端
func NewServer(addr string, log *zap.Logger, unary []grpc.UnaryServerInterceptor, stream []grpc.StreamServerInterceptor) (*Server, error) {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("listen %s failed: %w", addr, err)
	}
	srv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(unary...),
		grpc.ChainStreamInterceptor(stream...),
	)
	// 开启 reflection：便于 grpcurl/grpcui 调试
	reflection.Register(srv)
	return &Server{srv: srv, addr: addr, lis: lis, log: log}, nil
}

// GetServer 拿底层 grpc.Server，用于 RegisterService
func (s *Server) GetServer() *grpc.Server { return s.srv }

// Start 阻塞式 Serve
func (s *Server) Start() error {
	s.log.Info("gRPC server starting", zap.String("addr", s.addr))
	return s.srv.Serve(s.lis)
}

// Stop 优雅退出
func (s *Server) Stop(ctx context.Context) {
	s.log.Info("gRPC server stopping")
	done := make(chan struct{})
	go func() {
		s.srv.GracefulStop()
		close(done)
	}()
	select {
	case <-done:
	case <-ctx.Done():
		s.srv.Stop()
	}
}
