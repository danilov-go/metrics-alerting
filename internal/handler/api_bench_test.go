package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/danilov-go/metrics-alerting.git/internal/models"
	"github.com/danilov-go/metrics-alerting.git/internal/repository"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

const (
	mID  = "Alloc"
	mVal = 123.45
)

func BenchmarkApiUpdateHandler(b *testing.B) {
	logger := zaptest.NewLogger(b)
	storage := repository.InitMemStorage(repository.ConfigFile{}, logger.Sugar())
	h := NewMetricsHandler(storage, logger.Sugar())
	handler := h.ApiUpdateHandler()
	metric := models.Metrics{
		ID:    mID,
		MType: models.Gauge,
		Value: models.PointerFloat64(mVal),
	}
	body, err := json.Marshal(metric)
	assert.NoError(b, err)
	b.ReportAllocs()
	for b.Loop() {
		b.StopTimer()
		req := httptest.NewRequest(http.MethodPost, "/update", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		b.StartTimer()
		handler.ServeHTTP(w, req)
	}
}

func BenchmarkApiValueHandler(b *testing.B) {
	logger := zaptest.NewLogger(b)
	storage := repository.InitMemStorage(repository.ConfigFile{}, logger.Sugar())
	h := NewMetricsHandler(storage, logger.Sugar())
	handler := h.ApiValueHandler()
	ctx := b.Context()
	err := storage.SaveGauges(ctx, mID, mVal)
	assert.NoError(b, err)
	metric := models.Metrics{
		ID:    mID,
		MType: models.Gauge,
	}
	body, err := json.Marshal(metric)
	assert.NoError(b, err)
	b.ReportAllocs()
	for b.Loop() {
		b.StopTimer()
		req := httptest.NewRequest(http.MethodPost, "/value", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		b.StartTimer()
		handler.ServeHTTP(w, req)
	}
}

func BenchmarkApiUpdatesHandler(b *testing.B) {
	logger := zaptest.NewLogger(b)
	storage := repository.InitMemStorage(repository.ConfigFile{}, logger.Sugar())
	h := NewMetricsHandler(storage, logger.Sugar())
	handler := h.ApiUpdatesHandler()
	metrics := []models.Metrics{
		{ID: mID, MType: models.Gauge, Value: models.PointerFloat64(mVal)},
		{ID: "CPUutilization1", MType: models.Gauge, Value: models.PointerFloat64(38.88)},
		{ID: "CPUutilization2", MType: models.Gauge, Value: models.PointerFloat64(35.17)},
		{ID: "CPUutilization3", MType: models.Gauge, Value: models.PointerFloat64(22.22)},
		{ID: "CPUutilization4", MType: models.Gauge, Value: models.PointerFloat64(18.50)},
		{ID: "CPUutilization5", MType: models.Gauge, Value: models.PointerFloat64(0.99)},
		{ID: "CPUutilization6", MType: models.Gauge, Value: models.PointerFloat64(0.50)},
		{ID: "CPUutilization7", MType: models.Gauge, Value: models.PointerFloat64(0.49)},
		{ID: "CPUutilization8", MType: models.Gauge, Value: models.PointerFloat64(0.99)},
		{ID: "PollCount", MType: models.Counter, Delta: models.PointerInt64(5)},
	}
	body, err := json.Marshal(metrics)
	assert.NoError(b, err)
	b.ReportAllocs()
	for b.Loop() {
		b.StopTimer()
		req := httptest.NewRequest(http.MethodPost, "/updates", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		b.StartTimer()
		handler.ServeHTTP(w, req)
	}
}
