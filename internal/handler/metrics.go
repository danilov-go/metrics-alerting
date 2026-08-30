package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/danilov-go/metrics-alerting.git/internal/models"
	"github.com/go-chi/chi/v5"
)

// PostMetricsHandler возвращает обработчик для создания или обновления метрики через URL-параметры.
func (h *MetricsHandler) PostMetricsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		mType := chi.URLParam(r, "mType")
		mName := chi.URLParam(r, "mName")
		mVal := chi.URLParam(r, "mVal")
		if mName == "" {
			h.logger.Errorw("пустое имя метрики", "mType", mType)
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		switch mType {
		case models.Gauge:
			valueMetric, err := strconv.ParseFloat(mVal, 64)
			if err != nil {
				h.logger.Errorw("ошибка парсинга gauge", "error", err)
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}
			err = h.storage.SaveGauges(ctx, mName, valueMetric)
			if err != nil {
				h.logger.Errorw("ошибка сохранения gauge", "error", err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(http.StatusOK)
		case models.Counter:
			valueMetric, err := strconv.Atoi(mVal)
			if err != nil {
				h.logger.Errorw("ошибка парсинга counter", "error", err)
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}
			err = h.storage.SaveCounters(ctx, mName, int64(valueMetric))
			if err != nil {
				h.logger.Errorw("ошибка сохранения counter", "error", err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(http.StatusOK)
		default:
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
	}
}

// GetMetricHandler возвращает обработчик для получения текстового значения метрики по её типу и названию.
func (h *MetricsHandler) GetMetricHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		mType := chi.URLParam(r, "mType")
		mName := chi.URLParam(r, "mName")
		if mName == "" {
			h.logger.Errorw("пустое имя метрики", "mType", mType)
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		switch mType {
		case models.Gauge:
			val, err := h.storage.GetGauges(ctx, mName)
			if err != nil {
				h.logger.Errorw("ошибка получения gauge", "error", err)
				http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "%v", val)
		case models.Counter:
			val, err := h.storage.GetCounters(ctx, mName)
			if err != nil {
				h.logger.Errorw("ошибка получения сounter", "error", err)
				http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "%v", val)
		default:
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
	}
}

// ExposeMetricsHandler возвращает HTML-страницу со списком сохраненных метрик.
func (h *MetricsHandler) ExposeMetricsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		mapGauge, err := h.storage.GetAllGauges(ctx)
		if err != nil {
			h.logger.Errorw("ошибка получения gauge", "error", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		mapCounter, err := h.storage.GetAllCounters(ctx)
		if err != nil {
			h.logger.Errorw("ошибка получения counter", "error", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
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
