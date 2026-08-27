package handler_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/danilov-go/metrics-alerting.git/internal/handler"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestRequestLogger(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		method      string
		handlerFunc func(w http.ResponseWriter, r *http.Request)
		status      int64
		size        int64
	}{
		{
			name:   "GET запрос",
			url:    "/value/counter/PollCount",
			method: http.MethodGet,
			handlerFunc: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`5`))
			},
			status: http.StatusOK,
			size:   1,
		},
		{
			name:   "POST запрос",
			url:    "/update/gauge/Alloc/123.45",
			method: http.MethodPost,
			handlerFunc: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"id":"Alloc","type":"gauge","value":123.45}`))
			},
			status: http.StatusOK,
			size:   44,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core, logs := observer.New(zapcore.InfoLevel)
			testLogger := zap.New(core)
			middleware := handler.RequestLogger(testLogger)
			h := middleware(http.HandlerFunc(tt.handlerFunc))
			req := httptest.NewRequest(tt.method, tt.url, nil)
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)
			logEntry := logs.All()[0]
			message := logEntry.ContextMap()
			assert.Equal(t, tt.url, message["url"])
			assert.Equal(t, tt.method, message["method"])
			assert.Equal(t, tt.status, message["status"])
			assert.Equal(t, tt.size, message["size"])
			assert.Contains(t, message, "duration")
		})
	}
}
