package agent

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

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
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "text/plain; charset=utf-8", r.Header.Get("Content-Type"))
		path := r.URL.Path
		metric := strings.Split(path, "/")
		if len(metric) > 4 {
			name := metric[3]
			storage[name] = true
			switch metric[2] {
			case "gauge":
				_, err := strconv.ParseFloat(metric[4], 64)
				assert.NoError(t, err)
			case "counter":
				_, err := strconv.Atoi(metric[4])
				assert.NoError(t, err)
				assert.Equal(t, "PollCount", metric[3])
			default:
				t.Errorf("неизвестный тип метрики: %s", metric[1])
			}
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	u, err := url.Parse(server.URL)
	require.NoError(t, err)
	a := New(u.Host)
	var pollCount int64 = 4
	step := 5
	a.Run(&pollCount, step)
	assert.Equal(t, int64(5), pollCount)
	for _, v := range expMetric {
		assert.Contains(t, storage, v)
	}
}
