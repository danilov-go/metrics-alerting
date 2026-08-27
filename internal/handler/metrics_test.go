package handler_test

import (
	"fmt"
	"math/rand/v2"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/danilov-go/metrics-alerting.git/internal/handler"
	"github.com/danilov-go/metrics-alerting.git/internal/repository"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestPostMetricsHandler(t *testing.T) {
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
		name     string
		storage  *repository.MemStorage
		url      string
		expected want
	}{
		{
			name: "положительный тест gauge",
			url:  "/update/gauge/Alloc/123.45",
			expected: want{
				mExp: metricExpected{
					mType: "gauge",
					name:  "Alloc",
					valG:  123.45,
				},
				code: http.StatusOK,
			},
		},
		{
			name: "положительный тест counter",
			url:  "/update/counter/PollCount/15",
			expected: want{
				mExp: metricExpected{
					mType: "counter",
					name:  "PollCount",
					valC:  15,
				},
				code: http.StatusOK,
			},
		},
		{
			name: "запрос без имени и значения",
			url:  "/update/gauge/",
			expected: want{
				code: http.StatusNotFound,
			},
		},
		{
			name: "пустое имя метрики",
			url:  "/update/gauge//123.45",
			expected: want{
				code: http.StatusNotFound,
			},
		},
		{
			name: "неверное значение counter",
			url:  "/update/counter/PollCount/123.45",
			expected: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "неверное значение gauge",
			url:  "/update/gauge/Alloc/invalid",
			expected: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "неверный тип метрики",
			url:  "/update/invalid/Aloc/123,45",
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
			r.Post("/update/{mType}/{mName}/{mVal}", h.PostMetricsHandler())
			request := httptest.NewRequest(http.MethodPost, tt.url, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, request)
			assert.Equal(t, tt.expected.code, w.Code)
			switch tt.expected.mExp.mType {
			case "counter":
				assert.Contains(t, tt.storage.Counters, tt.expected.mExp.name)
				assert.Equal(t, tt.expected.mExp.valC, tt.storage.Counters[tt.expected.mExp.name])
			case "gauge":
				assert.Contains(t, tt.storage.Gauges, tt.expected.mExp.name)
				assert.Equal(t, tt.expected.mExp.valG, tt.storage.Gauges[tt.expected.mExp.name])
			}
		})
	}
}

func TestGetMetricHandler(t *testing.T) {
	type want struct {
		code int
		valG float64
		valC int64
	}
	tests := []struct {
		name     string
		mType    string
		mName    string
		valG     float64
		valC     int64
		expected want
	}{
		{
			name:  "положительный тест gauge",
			mType: "gauge",
			mName: "Alloc",
			valG:  123.45,
			expected: want{
				code: http.StatusOK,
				valG: 123.45,
			},
		},
		{
			name:  "положительный тест counter",
			mType: "counter",
			mName: "PollCount",
			valC:  15,
			expected: want{
				code: http.StatusOK,
				valC: 15,
			},
		},
		{
			name:  "неверный тип метрики",
			mType: "invalid",
			expected: want{
				code: http.StatusNotFound,
			},
		},
		{
			name:  "пустое имя метрики",
			mType: "counter",
			mName: "",
			expected: want{
				code: http.StatusNotFound,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := &repository.MemStorage{
				Gauges:   make(map[string]float64),
				Counters: make(map[string]int64),
			}
			if tt.mName != "" {
				switch tt.mType {
				case "counter":
					storage.SaveCounters(t.Context(), tt.mName, tt.valC)
				case "gauge":
					storage.SaveGauges(t.Context(), tt.mName, tt.valG)
				}
			}
			logger := zaptest.NewLogger(t)
			h := handler.NewMetricsHandler(storage, logger.Sugar())
			r := chi.NewRouter()
			r.Get("/value/{mType}/{mName}", h.GetMetricHandler())
			url := fmt.Sprintf("/value/%s/%s", tt.mType, tt.mName)
			request := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, request)
			assert.Equal(t, tt.expected.code, w.Code)
			if w.Code == http.StatusOK {
				switch tt.mType {
				case "counter":
					assert.Equal(t, fmt.Sprintf("%v", tt.expected.valC), w.Body.String())
				case "gauge":
					assert.Equal(t, fmt.Sprintf("%v", tt.expected.valG), w.Body.String())
				}
			}
		})
	}
}

func TestExposeMetricHandler(t *testing.T) {
	type want struct {
		code int
	}
	tests := []struct {
		name     string
		mName    []string
		expected want
	}{
		{
			name: "положительный тест ",
			mName: []string{
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
			},
			expected: want{
				code: http.StatusOK,
			},
		},
		{
			name:  "пустое хранилище",
			mName: []string{},
			expected: want{
				code: http.StatusOK,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := &repository.MemStorage{
				Gauges:   make(map[string]float64),
				Counters: make(map[string]int64),
			}
			for _, name := range tt.mName {
				if name == "PollCount" {
					storage.SaveCounters(t.Context(), name, rand.Int64())
				} else {
					storage.SaveGauges(t.Context(), name, rand.Float64())
				}
			}
			logger := zaptest.NewLogger(t)
			h := handler.NewMetricsHandler(storage, logger.Sugar())
			r := chi.NewRouter()
			r.Get("/", h.ExposeMetricsHandler())
			url := "/"
			request := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, request)
			assert.Equal(t, tt.expected.code, w.Code)
			assert.Equal(t, "text/html; charset=utf-8", w.Header().Get("Content-Type"))
			body := w.Body.String()
			if w.Code == http.StatusOK {
				for name, val := range storage.Gauges {
					assert.Contains(t, body, name)
					assert.Contains(t, body, fmt.Sprintf("%v", val))
				}
				for name, val := range storage.Counters {
					assert.Contains(t, body, name)
					assert.Contains(t, body, fmt.Sprintf("%v", val))
				}
			}
		})
	}
}
