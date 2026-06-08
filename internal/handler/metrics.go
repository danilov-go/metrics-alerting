package handler

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

func PostMetricsHandler(ctx context.Context, s Storage) http.HandlerFunc {
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
			s.SaveGauges(ctx, mName, valueMetric)
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(http.StatusOK)
		case "counter":
			valueMetric, err := strconv.Atoi(mVal)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			s.SaveCounters(ctx, mName, int64(valueMetric))
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(http.StatusOK)
		default:
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}
}

func GetMetricHandler(ctx context.Context, s Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mType := chi.URLParam(r, "mType")
		mName := chi.URLParam(r, "mName")
		if mName == "" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		switch mType {
		case "gauge":
			val, err := s.GetGauges(ctx, mName)
			if err != nil {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "%v", val)
		case "counter":
			val, err := s.GetCounters(ctx, mName)
			if err != nil {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "%v", val)
		default:
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}
}

func ExposeMetricsHandler(ctx context.Context, s Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mapGauge, err := s.GetAllGauges(ctx)
		if err != nil {
			http.Error(w, "ошибка получения gauge", http.StatusInternalServerError)
			return
		}
		mapCounter, err := s.GetAllCounters(ctx)
		if err != nil {
			http.Error(w, "ошибка получения counter", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if len(mapGauge) == 0 && len(mapCounter) == 0 {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusOK)
		for name, val := range mapGauge {
			fmt.Fprintf(w, "%v : %v<br>\n", name, val)
		}
		for name, val := range mapCounter {
			fmt.Fprintf(w, "%v : %v<br>\n", name, val)
		}
	}
}
