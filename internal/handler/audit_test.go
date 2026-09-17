package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/danilov-go/metrics-alerting.git/internal/audit"
	"github.com/danilov-go/metrics-alerting.git/internal/models"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

const (
	optionURL  = "URL"
	optionJSON = "JSON"
	addr       = "192.168.1.1"
)

var metrics = []models.Metrics{
	{ID: "Alloc", MType: models.Gauge, Value: models.Pointer(123.45)},
	{ID: "CPUutilization1", MType: models.Gauge, Value: models.Pointer(38.88)},
	{ID: "CPUutilization2", MType: models.Gauge, Value: models.Pointer(35.17)},
	{ID: "CPUutilization3", MType: models.Gauge, Value: models.Pointer(22.22)},
	{ID: "CPUutilization4", MType: models.Gauge, Value: models.Pointer(18.50)},
	{ID: "CPUutilization5", MType: models.Gauge, Value: models.Pointer(0.99)},
	{ID: "CPUutilization6", MType: models.Gauge, Value: models.Pointer(0.50)},
	{ID: "CPUutilization7", MType: models.Gauge, Value: models.Pointer(0.49)},
	{ID: "CPUutilization8", MType: models.Gauge, Value: models.Pointer(0.99)},
	{ID: "PollCount", MType: models.Counter, Delta: models.Pointer(int64(5))},
}

var metricsName = []string{
	"Alloc",
	"CPUutilization1",
	"CPUutilization2",
	"CPUutilization3",
	"CPUutilization4",
	"CPUutilization5",
	"CPUutilization6",
	"CPUutilization7",
	"CPUutilization8",
	"PollCount",
}

func TestAuditMiddleware(t *testing.T) {
	body, err := json.Marshal(metrics)
	require.NoError(t, err)
	type testCase struct {
		name     string
		body     []byte
		option   string
		status   int
		wantErr  bool
		expAudit audit.Audit
	}
	tests := []testCase{
		{
			name:    "положительный тест (JSON)",
			body:    body,
			option:  optionJSON,
			status:  http.StatusOK,
			wantErr: false,
			expAudit: audit.Audit{
				Metrics:   metricsName,
				IPAddress: addr,
			},
		},
		{
			name:    "положительный тест (URL)",
			body:    nil,
			option:  optionURL,
			status:  http.StatusOK,
			wantErr: false,
			expAudit: audit.Audit{
				Metrics:   []string{"Alloc"},
				IPAddress: addr,
			},
		},
		{
			name:    "ошибка хендлера",
			body:    body,
			option:  optionJSON,
			status:  http.StatusBadRequest,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zaptest.NewLogger(t).Sugar()
			event := audit.NewEvent(logger)
			path := filepath.Join(t.TempDir(), "audit.log")
			sub := audit.NewFileSubscriber(path, logger)
			event.Register(sub)
			req := httptest.NewRequest(http.MethodPost, "/update", bytes.NewReader(tt.body))
			req.RemoteAddr = addr
			switch tt.option {
			case optionURL:
				r := chi.NewRouteContext()
				r.URLParams.Add("mName", "Alloc")
				r.URLParams.Add("mType", models.Gauge)
				r.URLParams.Add("mVal", "123.45")
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, r))
			case optionJSON:
				req.Header.Set("Content-Type", "application/json")
			default:
				t.Fatalf("неизвестная опция: %s", tt.option)
			}
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.status)
			})
			h := AuditMiddleware(event)(next)
			h.ServeHTTP(httptest.NewRecorder(), req)
			assert.NoError(t, sub.Close())
			if tt.wantErr {
				info, err := os.Stat(path)
				if err == nil {
					assert.Zero(t, info.Size())
				}
				return
			}
			data, err := os.ReadFile(path)
			require.NoError(t, err)
			require.NotEmpty(t, data)
			var audit audit.Audit
			require.NoError(t, json.Unmarshal(data, &audit))
			assert.NotZero(t, audit.TS)
			tt.expAudit.TS = audit.TS
			assert.Equal(t, tt.expAudit, audit)
		})
	}
}
