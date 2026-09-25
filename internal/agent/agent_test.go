package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/danilov-go/metrics-alerting.git/internal/config"
	"github.com/danilov-go/metrics-alerting.git/internal/models"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
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
		"TotalMemory",
		"FreeMemory",
	}
	cores := runtime.NumCPU()
	for i := 0; i < cores; i++ {
		expMetric = append(expMetric, fmt.Sprintf("CPUutilization%d", i+1))
	}
	var storage sync.Map
	r := chi.NewRouter()
	r.Post("/updates/", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "gzip", r.Header.Get("Content-Encoding"))
		var metric []models.Metrics
		var buf bytes.Buffer
		wg, err := gzip.NewReader(r.Body)
		assert.NoError(t, err)
		defer func() { _ = wg.Close() }()
		_, err = buf.ReadFrom(wg)
		assert.NoError(t, err)
		err = json.Unmarshal(buf.Bytes(), &metric)
		assert.NoError(t, err)
		for _, m := range metric {
			storage.Store(m.ID, true)
			switch m.MType {
			case models.Gauge:
				assert.NotNil(t, m.Value)
			case models.Counter:
				assert.NotNil(t, m.Delta)
			default:
				t.Errorf("неизвестный тип метрики: %s", m.MType)
			}
		}
		w.WriteHeader(http.StatusOK)
	})
	server := httptest.NewServer(r)
	defer server.Close()
	u, err := url.Parse(server.URL)
	require.NoError(t, err)
	logger := zaptest.NewLogger(t)
	cfg := config.ConfigAgent{
		Net: config.NetAddress{
			Host: u.Hostname(),
			Port: 8080,
		},
		RateLimit:      2,
		PollInterval:   1,
		ReportInterval: 1,
		Key:            "",
	}
	a, err := New(cfg, logger.Sugar())
	require.NoError(t, err)
	if httpSender, ok := a.Sender.(*HTTPSender); ok {
		httpSender.client.SetBaseURL(server.URL)
		httpSender.client.SetTimeout(5 * time.Second)
	}
	assert.NoError(t, err)
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(t.Context())
	assert.NoError(t, err)
	go a.Run(ctx, nil)
	time.Sleep(3 * time.Second)
	cancel()
	wg.Wait()
	for _, v := range expMetric {
		_, ok := storage.Load(v)
		assert.True(t, ok)
	}
}

func TestNew(t *testing.T) {
	logger := zaptest.NewLogger(t)
	tests := []struct {
		name    string
		cfg     config.ConfigAgent
		option  string
		wantErr bool
	}{
		{
			name: "положительный тест (HTTP)",
			cfg: config.ConfigAgent{
				PollInterval:   2,
				ReportInterval: 10,
				RateLimit:      3,
				Key:            "secret_key",
				Net: config.NetAddress{
					Host: "localhost",
					Port: 8080,
				},
			},
			option:  "http",
			wantErr: false,
		},
		{
			name: "положительный тест (gRPC)",
			cfg: config.ConfigAgent{
				GrpcAddress:    "127.0.0.1:8082",
				PollInterval:   1,
				ReportInterval: 5,
			},
			option:  "grpc",
			wantErr: false,
		},
		{
			name: "ошибка инициализации gRPC",
			cfg: config.ConfigAgent{
				GrpcAddress: "unknow%",
			},
			option:  "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, err := New(tt.cfg, logger.Sugar())
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, a)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, a)
			assert.Equal(t, int(tt.cfg.PollInterval), a.pollInterval)
			assert.Equal(t, int(tt.cfg.ReportInterval), a.reportInterval)
			assert.Equal(t, tt.cfg.RateLimit, a.rateLimit)
			assert.NotNil(t, a.Sender)
			switch tt.option {
			case "http":
				_, ok := a.Sender.(*HTTPSender)
				assert.True(t, ok)
			case "grpc":
				_, ok := a.Sender.(*GRPCSender)
				assert.True(t, ok)
				if gSender, ok := a.Sender.(*GRPCSender); ok {
					_ = gSender.Close()
				}
			}
		})
	}
}
