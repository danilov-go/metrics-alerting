package handler

import (
	"context"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func UnaryInterceptor(ipNet *net.IPNet) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if ipNet == nil {
			return handler(ctx, req)
		}
		var ipStr string
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			values := md.Get("x-real-ip")
			if len(values) > 0 {
				ipStr = values[0]
			}
		}
		if len(ipStr) == 0 {
			return nil, status.Error(codes.PermissionDenied, "не задан заголовок x-real-ip")
		}
		ip := net.ParseIP(ipStr)
		if ip == nil {
			return nil, status.Error(codes.PermissionDenied, "некорректный IP-адрес")
		}
		if !ipNet.Contains(ip) {
			return nil, status.Error(codes.PermissionDenied, "нет прав доступа")
		}
		return handler(ctx, req)
	}
}
