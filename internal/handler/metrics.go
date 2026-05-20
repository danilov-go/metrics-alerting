package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/danilov-go/metrics-alerting.git/internal/repository"

	"github.com/go-chi/chi/v5"
)

func PostMetricsHandler(s *repository.MemStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mType := chi.URLParam(r, "mType")
		mName := chi.URLParam(r, "mName")
		mVal := chi.URLParam(r, "mVal")
		if mName == "" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		switch mType {
		case "gauge":
			valueMetric, err := strconv.ParseFloat(mVal, 64)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			s.SaveGauges(mName, valueMetric)
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(http.StatusOK)
		case "counter":
			valueMetric, err := strconv.Atoi(mVal)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			s.SaveCounters(mName, int64(valueMetric))
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(http.StatusOK)
		default:
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}
}

func GetMetricHandler(s *repository.MemStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mType := chi.URLParam(r, "mType")
		mName := chi.URLParam(r, "mName")
		if mName == "" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		switch mType {
		case "gauge":
			val, valid := s.GetGauges(mName)
			if !valid {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "%f", val)
		case "counter":
			val, valid := s.GetCounters(mName)
			if !valid {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "%d", val)
		default:
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}
}

func ExposeMetricsHandler(s *repository.MemStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mapGauge := s.GetAllGauges()
		mapCounter := s.GetAllCounters()
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if len(mapGauge) == 0 && len(mapCounter) == 0 {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusOK)
		for name, val := range mapGauge {
			fmt.Fprintf(w, "%s : %f<br>\n", name, val)
		}
		for name, val := range mapCounter {
			fmt.Fprintf(w, "%s : %d<br>\n", name, val)
		}
	}
}
