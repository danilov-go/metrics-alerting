package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/danilov-go/metrics-alerting.git/internal/handler"
	"github.com/danilov-go/metrics-alerting.git/internal/models"
	"github.com/danilov-go/metrics-alerting.git/internal/repository"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestAPIUpdateHandler(t *testing.T) {
	type metricExpected struct {
		mType string
		name  string
		valG  float64
		valC  int64
	}
	type want struct {
		mExp metricExpected
		code int
	}

	tests := []struct {
		name        string
		contentType string
		storage     *repository.MemStorage
		m           models.Metrics
		rawBody     string
		expected    want
	}{
		{
			name:        "положительный тест gauge",
			contentType: "application/json",
			m: models.Metrics{
				ID:    "Alloc",
				MType: models.Gauge,
				Value: models.PointerFloat64(123.45),
			},
			expected: want{
				mExp: metricExpected{
					mType: models.Gauge,
					name:  "Alloc",
					valG:  123.45,
				},
				code: http.StatusOK,
			},
		},
		{
			name:        "положительный тест counter",
			contentType: "application/json",
			m: models.Metrics{
				ID:    "PollCount",
				MType: models.Counter,
				Delta: models.PointerInt64(15),
			},
			expected: want{
				mExp: metricExpected{
					mType: models.Counter,
					name:  "PollCount",
					valC:  15,
				},
				code: http.StatusOK,
			},
		},
		{
			name:        "пустое имя метрики ID",
			contentType: "application/json",
			m: models.Metrics{
				ID:    "",
				MType: models.Gauge,
				Value: models.PointerFloat64(123.45),
			},
			expected: want{
				code: http.StatusNotFound,
			},
		},
		{
			name:        "неверный тип метрики",
			contentType: "application/json",
			m: models.Metrics{
				ID:    "Alloc",
				MType: "invalid_type",
				Value: models.PointerFloat64(123.45),
			},
			expected: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name:        "некорректный JSON",
			contentType: "application/json",
			rawBody:     "{error json}",
			expected: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name:        "отсутствует value",
			contentType: "application/json",
			m: models.Metrics{
				ID:    "Alloc",
				MType: models.Gauge,
				Value: nil,
			},
			expected: want{
				code: http.StatusNotFound,
			},
		},
		{
			name:        "отсутствует delta",
			contentType: "application/json",
			m: models.Metrics{
				ID:    "PollCount",
				MType: models.Counter,
				Delta: nil,
			},
			expected: want{
				code: http.StatusNotFound,
			},
		},
		{
			name:        "неверный Content-Type",
			rawBody:     "[]",
			contentType: "text/plain",
			expected: want{
				code: http.StatusBadRequest,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.storage = &repository.MemStorage{
				Gauges:   make(map[string]float64),
				Counters: make(map[string]int64),
			}
			logger := zaptest.NewLogger(t)
			h := handler.NewMetricsHandler(tt.storage, logger.Sugar())
			r := chi.NewRouter()
			r.Post("/update", h.APIUpdateHandler())
			var body []byte
			var err error
			if tt.rawBody != "" {
				body = []byte(tt.rawBody)
			} else {
				body, err = json.Marshal(tt.m)
				assert.NoError(t, err)
			}
			buf := bytes.NewBuffer(body)
			request := httptest.NewRequest(http.MethodPost, "/update", buf)
			request.Header.Set("Content-Type", tt.contentType)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, request)
			assert.Equal(t, tt.expected.code, w.Code)
			if tt.expected.code == http.StatusOK {
				switch tt.expected.mExp.mType {
				case models.Counter:
					assert.Contains(t, tt.storage.Counters, tt.expected.mExp.name)
					assert.Equal(t, tt.expected.mExp.valC, tt.storage.Counters[tt.expected.mExp.name])
				case models.Gauge:
					assert.Contains(t, tt.storage.Gauges, tt.expected.mExp.name)
					assert.Equal(t, tt.expected.mExp.valG, tt.storage.Gauges[tt.expected.mExp.name])
				}
			}
		})
	}
}

func TestAPIValueHandler(t *testing.T) {
	type want struct {
		code  int
		mType string
		mName string
		valG  float64
		valC  int64
	}
	tests := []struct {
		name        string
		rawBody     string
		m           models.Metrics
		expected    want
		contentType string
	}{
		{
			name: "положительный тест gauge",
			m: models.Metrics{
				ID:    "Alloc",
				MType: models.Gauge,
			},
			expected: want{
				code:  http.StatusOK,
				mType: models.Gauge,
				mName: "Alloc",
				valG:  123.45,
			},
			contentType: "application/json",
		},
		{
			name: "положительный тест counter",
			m: models.Metrics{
				ID:    "PollCount",
				MType: models.Counter,
			},
			expected: want{
				code:  http.StatusOK,
				mType: models.Counter,
				mName: "PollCount",
				valC:  15,
			},
			contentType: "application/json",
		},
		{
			name: "неверный тип метрики",
			m: models.Metrics{
				ID:    "Alloc",
				MType: "invalid",
			},
			expected: want{
				code: http.StatusBadRequest,
			},
			contentType: "application/json",
		},
		{
			name: "пустое имя метрики",
			m: models.Metrics{
				ID:    "",
				MType: models.Counter,
			},
			expected: want{
				code: http.StatusNotFound,
			},
			contentType: "application/json",
		},
		{
			name: "неверный Content-Type",
			m: models.Metrics{
				ID:    "Alloc",
				MType: models.Gauge,
			},
			expected: want{
				code: http.StatusBadRequest,
			},
			contentType: "text/plain",
		},
		{
			name: "метрика gauge отсутствует в хранилище",
			m: models.Metrics{
				ID:    "unknow",
				MType: models.Gauge,
			},
			expected: want{
				code: http.StatusNotFound,
			},
			contentType: "application/json",
		},
		{
			name: "метрика counter отсутствует в хранилище",
			m: models.Metrics{
				ID:    "unknow",
				MType: models.Counter,
			},
			expected: want{
				code: http.StatusNotFound,
			},
			contentType: "application/json",
		},
		{
			name:        "некорректный JSON",
			rawBody:     "{error json}",
			contentType: "application/json",
			expected: want{
				code: http.StatusBadRequest,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := &repository.MemStorage{
				Gauges:   make(map[string]float64),
				Counters: make(map[string]int64),
			}
			if tt.expected.mName != "" {
				switch tt.expected.mType {
				case models.Counter:
					err := storage.SaveCounters(t.Context(), tt.expected.mName, tt.expected.valC)
					assert.NoError(t, err)
				case models.Gauge:
					err := storage.SaveGauges(t.Context(), tt.expected.mName, tt.expected.valG)
					assert.NoError(t, err)
				}
			}
			logger := zaptest.NewLogger(t)
			h := handler.NewMetricsHandler(storage, logger.Sugar())
			r := chi.NewRouter()
			r.Post("/value", h.APIValueHandler())
			var body []byte
			var err error
			if tt.rawBody != "" {
				body = []byte(tt.rawBody)
			} else {
				body, err = json.Marshal(tt.m)
				assert.NoError(t, err)
			}

			assert.NoError(t, err)
			buf := bytes.NewBuffer(body)
			request := httptest.NewRequest(http.MethodPost, "/value", buf)
			request.Header.Set("Content-Type", tt.contentType)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, request)
			assert.Equal(t, tt.expected.code, w.Code)
			if w.Code == http.StatusOK {
				var respMetrics models.Metrics
				err := json.Unmarshal(w.Body.Bytes(), &respMetrics)
				assert.NoError(t, err)
				switch tt.expected.mType {
				case models.Counter:
					assert.NotNil(t, respMetrics.Delta)
					assert.Equal(t, tt.expected.valC, *respMetrics.Delta)
				case models.Gauge:
					assert.NotNil(t, respMetrics.Value)
					assert.Equal(t, tt.expected.valG, *respMetrics.Value)
				}
			}
		})
	}
}

func TestAPIUpdatesHandler(t *testing.T) {
	type metricExpected struct {
		mType string
		name  string
		valG  float64
		valC  int64
	}
	type want struct {
		mExp metricExpected
		code int
	}

	tests := []struct {
		name        string
		storage     *repository.MemStorage
		m           models.Metrics
		contentType string
		rawBody     string
		expected    want
	}{
		{
			name: "положительный тест gauge",
			m: models.Metrics{
				ID:    "Alloc",
				MType: models.Gauge,
				Value: models.PointerFloat64(123.45),
			},
			contentType: "application/json",
			expected: want{
				mExp: metricExpected{
					mType: models.Gauge,
					name:  "Alloc",
					valG:  123.45,
				},
				code: http.StatusOK,
			},
		},
		{
			name: "положительный тест counter",
			m: models.Metrics{
				ID:    "PollCount",
				MType: models.Counter,
				Delta: models.PointerInt64(15),
			},
			contentType: "application/json",
			expected: want{
				mExp: metricExpected{
					mType: models.Counter,
					name:  "PollCount",
					valC:  15,
				},
				code: http.StatusOK,
			},
		},
		{
			name: "неверный тип метрики",
			m: models.Metrics{
				ID:    "Alloc",
				MType: "invalid_type",
				Value: models.PointerFloat64(123.45),
			},
			contentType: "application/json",
			expected: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name:        "некорректный JSON",
			rawBody:     "{error json}",
			contentType: "application/json",
			expected: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name:        "неверный Content-Type",
			rawBody:     "[]",
			contentType: "text/plain",
			expected: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "пустой Delta",
			m: models.Metrics{
				ID:    "PollCount",
				MType: models.Counter,
				Delta: nil,
			},
			contentType: "application/json",
			expected: want{
				code: http.StatusInternalServerError,
			},
		},
		{
			name: "пустой Value",
			m: models.Metrics{
				ID:    "Alloc",
				MType: models.Gauge,
				Value: nil,
			},
			contentType: "application/json",
			expected: want{
				code: http.StatusInternalServerError,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.storage = &repository.MemStorage{
				Gauges:   make(map[string]float64),
				Counters: make(map[string]int64),
			}
			logger := zaptest.NewLogger(t)
			h := handler.NewMetricsHandler(tt.storage, logger.Sugar())
			r := chi.NewRouter()
			r.Post("/updates", h.APIUpdatesHandler())
			var body []byte
			var err error
			if tt.rawBody != "" {
				body = []byte(tt.rawBody)
			} else {
				body, err = json.Marshal([]models.Metrics{tt.m})
				assert.NoError(t, err)
			}
			buf := bytes.NewBuffer(body)
			request := httptest.NewRequest(http.MethodPost, "/updates", buf)
			request.Header.Set("Content-Type", tt.contentType)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, request)
			assert.Equal(t, tt.expected.code, w.Code)
			if tt.expected.code == http.StatusOK {
				switch tt.expected.mExp.mType {
				case models.Counter:
					assert.Contains(t, tt.storage.Counters, tt.expected.mExp.name)
					assert.Equal(t, tt.expected.mExp.valC, tt.storage.Counters[tt.expected.mExp.name])
				case models.Gauge:
					assert.Contains(t, tt.storage.Gauges, tt.expected.mExp.name)
					assert.Equal(t, tt.expected.mExp.valG, tt.storage.Gauges[tt.expected.mExp.name])
				}
			}
		})
	}
}
