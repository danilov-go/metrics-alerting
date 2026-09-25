package handler_test

import (
	"context"
	"net"
	"testing"

	"github.com/danilov-go/metrics-alerting.git/internal/handler"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestUnaryInterceptor(t *testing.T) {
	_, trustedNet, err := net.ParseCIDR("192.168.1.0/24")
	require.NoError(t, err)
	h := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "test_response", nil
	}
	tests := []struct {
		name       string
		ipNet      *net.IPNet
		ctx        context.Context
		expError   bool
		code       codes.Code
		expMessage string
	}{
		{
			name:     "IP не задан",
			ipNet:    nil,
			ctx:      context.Background(),
			expError: false,
		},
		{
			name:     "положительный тест",
			ipNet:    trustedNet,
			ctx:      metadata.NewIncomingContext(context.Background(), metadata.Pairs("x-real-ip", "192.168.1.5")),
			expError: false,
		},
		{
			name:       "заголовок x-real-ip отсутствует",
			ipNet:      trustedNet,
			ctx:        context.Background(),
			expError:   true,
			code:       codes.PermissionDenied,
			expMessage: "не задан заголовок x-real-ip",
		},
		{
			name:       "пустой заголовок x-real-ip",
			ipNet:      trustedNet,
			ctx:        metadata.NewIncomingContext(context.Background(), metadata.Pairs("x-real-ip", "")),
			expError:   true,
			code:       codes.PermissionDenied,
			expMessage: "не задан заголовок x-real-ip",
		},
		{
			name:       "некорректный IP-адреса",
			ipNet:      trustedNet,
			ctx:        metadata.NewIncomingContext(context.Background(), metadata.Pairs("x-real-ip", "not-an-ip-string")),
			expError:   true,
			code:       codes.PermissionDenied,
			expMessage: "некорректный IP-адрес",
		},
		{
			name:       "нет прав доступа",
			ipNet:      trustedNet,
			ctx:        metadata.NewIncomingContext(context.Background(), metadata.Pairs("x-real-ip", "193.168.1.5")),
			expError:   true,
			code:       codes.PermissionDenied,
			expMessage: "нет прав доступа",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interceptor := handler.UnaryInterceptor(tt.ipNet)
			resp, err := interceptor(tt.ctx, "test", &grpc.UnaryServerInfo{}, h)
			if !tt.expError {
				assert.NoError(t, err)
				assert.Equal(t, "test_response", resp)
			} else {
				assert.Nil(t, resp)
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.code, st.Code())
				assert.Equal(t, tt.expMessage, st.Message())
			}
		})
	}
}
