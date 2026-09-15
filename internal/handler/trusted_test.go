package handler

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTrustedMiddleware(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	tests := []struct {
		name          string
		trustedSubnet string
		addr          string
		wantStatus    int
	}{
		{
			name:          "положительный тест",
			trustedSubnet: "192.168.1.0/24",
			addr:          "192.168.1.5",
			wantStatus:    http.StatusOK,
		},
		{
			name:          "адрес не входит в доверенную подсеть",
			trustedSubnet: "192.168.1.0/24",
			addr:          "193.168.1.5",
			wantStatus:    http.StatusForbidden,
		},
		{
			name:          "отсутствует заголовок X-Real-IP",
			trustedSubnet: "192.168.1.0/24",
			addr:          "",
			wantStatus:    http.StatusForbidden,
		},
		{
			name:          "невалидный адрес",
			trustedSubnet: "192.168.1.0/24",
			addr:          "unknow",
			wantStatus:    http.StatusForbidden,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, ipNet, err := net.ParseCIDR(tt.trustedSubnet)
			require.NoError(t, err)
			middleware := TrustedMiddleware(ipNet)
			h := middleware(handler)
			req := httptest.NewRequest(http.MethodPost, "/metrics", nil)
			if tt.addr != "" {
				req.Header.Set("X-Real-IP", tt.addr)
			}
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)
			assert.Equal(t, rec.Code, tt.wantStatus)
		})
	}
}
