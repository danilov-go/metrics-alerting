package server

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestServer(t *testing.T) {
	tests := []struct {
		name    string
		port    string
		wantErr bool
	}{
		{
			name:    "положительный тест",
			port:    "localhost:8080",
			wantErr: false,
		},
		{
			name:    "неверный адрес",
			port:    "invalid port",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zaptest.NewLogger(t)
			serv := New(tt.port, logger.Sugar(), http.NewServeMux())
			if tt.wantErr {
				err := serv.Run()
				assert.Error(t, err)
			} else {
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
