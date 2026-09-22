package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
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
	a := New(cfg, logger.Sugar())
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

func TestAgent_Encrypt(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	publicKey := &privateKey.PublicKey
	tests := []struct {
		name      string
		publicKey *rsa.PublicKey
		body      []byte
	}{
		{
			name:      "положительный тест",
			publicKey: publicKey,
			body:      []byte(""),
		},
		{
			name:      "положительный тест",
			publicKey: publicKey,
			body:      []byte(`[{"id":"Alloc","type":"gauge","value":123.45}]`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cipherBody, cipherKey, err := encrypt(tt.publicKey, tt.body)
			assert.NoError(t, err)
			assert.NotEmpty(t, cipherBody)
			assert.NotEmpty(t, cipherKey)
			aesKey, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, privateKey, cipherKey, nil)
			require.NoError(t, err)
			assert.Len(t, aesKey, 32)
			block, err := aes.NewCipher(aesKey)
			require.NoError(t, err)
			gsm, err := cipher.NewGCM(block)
			require.NoError(t, err)
			nonceSize := gsm.NonceSize()
			require.GreaterOrEqual(t, len(cipherBody), nonceSize)
			nonce := cipherBody[:nonceSize]
			body := cipherBody[nonceSize:]
			decryptedBody, err := gsm.Open(nil, nonce, body, nil)
			require.NoError(t, err)
			assert.Equal(t, string(tt.body), string(decryptedBody))

		})
	}
}
