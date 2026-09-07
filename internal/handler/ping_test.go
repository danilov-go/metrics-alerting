package handler_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/danilov-go/metrics-alerting.git/internal/handler"
	"github.com/danilov-go/metrics-alerting.git/internal/repository"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestPingHandler(t *testing.T) {
	logger := zaptest.NewLogger(t).Sugar()
	cfg := repository.ConfigFile{
		Path:     "",
		Interval: 0,
		Restore:  false,
	}
	storage := repository.InitMemStorage(cfg, logger)
	h := handler.NewMetricsHandler(storage, logger)
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	w := httptest.NewRecorder()
	handlerFunc := h.PingHandler()
	handlerFunc.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
