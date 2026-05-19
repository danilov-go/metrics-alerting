package handler_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/danilov-go/metrics-alerting.git/internal/handler"
	"github.com/danilov-go/metrics-alerting.git/internal/repository"
	"github.com/stretchr/testify/assert"
)

func TestMetricsHandler(t *testing.T) {
	type metricExpected struct {
		mType string
		name  string
		valG  float64
		valC  int64
	}
	type want struct {
		mExp metricExpected
		url  string
		сode int
	}
	tests := []struct {
		name     string
		storage  *repository.MemStorage
		expected want
	}{
		{
			name: "положительный тест gauge",
			expected: want{
				url: "/update/gauge/Alloc/123.45",
				mExp: metricExpected{
					mType: "gauge",
					name:  "Alloc",
					valG:  123.45,
				},
				сode: http.StatusOK,
			},
		},
		{
			name: "положительный тест counter",
			expected: want{
				url: "/update/counter/PollCount/15",
				mExp: metricExpected{
					mType: "counter",
					name:  "PollCount",
					valC:  15,
				},
				сode: http.StatusOK,
			},
		},
		{
			name: "запрос без имени и значения",
			expected: want{
				url:  "/update/gauge/",
				сode: http.StatusNotFound,
			},
		},
		{
			name: "пустое имя метрики",
			expected: want{
				url:  "/update/gauge//123.45",
				сode: http.StatusNotFound,
			},
		},
		{
			name: "неверное значение counter",
			expected: want{
				url:  "/update/counter/PollCount/123.45",
				сode: http.StatusBadRequest,
			},
		},
		{
			name: "неверное значение gauge",
			expected: want{
				url:  "/update/gauge/Alloc/invalid",
				сode: http.StatusBadRequest,
			},
		},
		{
			name: "неверный тип метрики",
			expected: want{
				url:  "/update/invalid/Aloc/123,45",
				сode: http.StatusBadRequest,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.storage = &repository.MemStorage{
				Gauges:   make(map[string]float64),
				Counters: make(map[string]int64),
			}
			request := httptest.NewRequest(http.MethodPost, tt.expected.url, nil)
			w := httptest.NewRecorder()
			h := handler.MetricsHandler(tt.storage)
			h.ServeHTTP(w, request)
			assert.Equal(t, tt.expected.сode, w.Code)
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
