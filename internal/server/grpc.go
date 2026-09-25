package server

import (
	"net"

	"google.golang.org/grpc"
)

// GRPCServer определяет конфигурацию gRPC-сервера.
type GRPCServer struct {
	Server   *grpc.Server
	Listener net.Listener
	Logger   log
	Address  string
}

// NewGRPC создает новый экземпляр gRPC-сервера.
func NewGRPC(address string, l log, s *grpc.Server, listen net.Listener) *GRPCServer {
	return &GRPCServer{
		Server:   s,
		Listener: listen,
		Logger:   l,
		Address:  address,
	}
}

// Run запускает gRPC-сервер.
func (serv *GRPCServer) Run() error {
	if err := serv.Server.Serve(serv.Listener); err != nil && err != grpc.ErrServerStopped {
		return err
	}
	return nil
}

// Stop останавливает gRPC-сервер.
func (serv *GRPCServer) Stop() error {
	serv.Server.GracefulStop()
	return nil
}
