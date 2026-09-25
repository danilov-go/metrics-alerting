package server

import (
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
	"google.golang.org/grpc"
)

func TestGRPCServer(t *testing.T) {
	tests := []struct {
		name    string
		addr    string
		wantErr bool
	}{
		{
			name:    "положительный тест",
			addr:    "localhost:8080",
			wantErr: false,
		},
		{
			name:    "неверный адрес",
			addr:    "invalid-addr",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zaptest.NewLogger(t)
			baseGRPCServer := grpc.NewServer()
			listen, err := net.Listen("tcp", tt.addr)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if listen != nil {
					defer listen.Close()
				}
				serv := NewGRPC(tt.addr, logger.Sugar(), baseGRPCServer, listen)
				go func() {
					_ = serv.Run()
				}()
				time.Sleep(5 * time.Millisecond)
				err := serv.Stop()
				assert.NoError(t, err)
			}
		})
	}
}
