package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/danilov-go/metrics-alerting.git/internal/repository"
)

func MetricsHandler(m *repository.MemStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		path := r.URL.Path
		metric := strings.Split(path, "/")
		if len(metric) < 5 {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if metric[3] == "" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		switch metric[2] {
		case "gauge":
			valueMetric, err := strconv.ParseFloat(metric[4], 64)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			m.SaveGauges(metric[3], valueMetric)
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(http.StatusOK)
		case "counter":
			valueMetric, err := strconv.Atoi(metric[4])
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			m.SaveCounters(metric[3], int64(valueMetric))
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(http.StatusOK)
		default:
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}
}
