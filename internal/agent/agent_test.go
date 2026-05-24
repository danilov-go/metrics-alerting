package agent

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"github.com/danilov-go/metrics-alerting.git/internal/logger"
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
	r.Post("/update/{mType}/{mName}/{mVal}", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "text/plain; charset=utf-8", r.Header.Get("Content-Type"))
		mType := chi.URLParam(r, "mType")
		mName := chi.URLParam(r, "mName")
		mVal := chi.URLParam(r, "mVal")
		storage[mName] = true
		switch mType {
		case "gauge":
			_, err := strconv.ParseFloat(mVal, 64)
			assert.NoError(t, err)
		case "counter":
			_, err := strconv.Atoi(mVal)
			assert.NoError(t, err)
			assert.Equal(t, "PollCount", mName)
		default:
			t.Errorf("неизвестный тип метрики: %s", mType)
		}
		w.WriteHeader(http.StatusOK)
	})
	server := httptest.NewServer(r)
	defer server.Close()
	u, err := url.Parse(server.URL)
	require.NoError(t, err)
	a := New(u.Host, logger.Log.Sugar())
	var pollCount int64 = 4
	step := 5
	a.Run(&pollCount, int64(step))
	assert.Equal(t, int64(5), pollCount)
	for _, v := range expMetric {
		assert.Contains(t, storage, v)
	}
}
