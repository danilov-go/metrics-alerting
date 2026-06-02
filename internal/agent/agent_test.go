package agent

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"

	"github.com/danilov-go/metrics-alerting.git/internal/logger"
	"github.com/danilov-go/metrics-alerting.git/internal/models"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAgent_Run(t *testing.T) {
	expMetric := []string{
		"Alloc",
		"BuckHashSys",
		"Frees",
		"GCCPUFraction",
		"GCSys",
		"HeapAlloc",
		"HeapIdle",
		"HeapInuse",
		"HeapObjects",
		"HeapReleased",
		"HeapSys",
		"LastGC",
		"Lookups",
		"MCacheInuse",
		"MCacheSys",
		"MSpanInuse",
		"MSpanSys",
		"Mallocs",
		"NextGC",
		"NumForcedGC",
		"NumGC",
		"OtherSys",
		"PauseTotalNs",
		"StackInuse",
		"StackSys",
		"Sys",
		"TotalAlloc",
		"RandomValue",
		"PollCount",
	}
	storage := make(map[string]bool)
	r := chi.NewRouter()
	r.Post("/update/", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "gzip", r.Header.Get("Content-Encoding"))
		var m models.Metrics
		var buf bytes.Buffer
		wg, err := gzip.NewReader(r.Body)
		assert.NoError(t, err)
		defer wg.Close()
		_, err = buf.ReadFrom(wg)
		assert.NoError(t, err)
		err = json.Unmarshal(buf.Bytes(), &m)
		assert.NoError(t, err)
		storage[m.ID] = true
		switch m.MType {
		case models.Gauge:
			assert.NotNil(t, m.Value)
		case models.Counter:
			assert.NotNil(t, m.Delta)
		default:
			t.Errorf("неизвестный тип метрики: %s", m.MType)
		}
		w.WriteHeader(http.StatusOK)
	})
	server := httptest.NewServer(r)
	defer server.Close()
	u, err := url.Parse(server.URL)
	require.NoError(t, err)
	err = logger.Initialize("info")
	require.NoError(t, err)
	a := New(u.Host, logger.Log.Sugar())
	var pollCount atomic.Int64
	pollCount.Store(4)
	pollCount.Add(1)
	metrics := a.Get(pollCount.Load())
	a.Run(metrics)
	assert.Equal(t, int64(5), pollCount.Load())
	for _, v := range expMetric {
		assert.Contains(t, storage, v)
	}
}
