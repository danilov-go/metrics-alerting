package handler

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"sync"

	"net/http"

	"github.com/danilov-go/metrics-alerting.git/internal/models"
)

var pool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

// ApiUpdateHandler возвращает обработчик для обновления или создания одиночной метрики из JSON.
func (h *MetricsHandler) ApiUpdateHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		var m models.Metrics
		buf := pool.Get().(*bytes.Buffer)
		buf.Reset()
		defer pool.Put(buf)
		if !strings.Contains(r.Header.Get("Content-Type"), "application/json") {
			h.logger.Errorw("неверный контент тип", "Content-Type", r.Header.Get("Content-Type"))
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		if _, err := buf.ReadFrom(r.Body); err != nil {
			h.logger.Errorw("ошибка чтения тела запроса", "error", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
			h.logger.Errorw("ошибка десериализации", "error", err)
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		if m.ID == "" {
			h.logger.Errorw("пустое имя метрики", "mType", m.MType)
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		switch m.MType {
		case models.Gauge:
			if m.Value == nil {
				h.logger.Errorw("значение value nil", "ID", m.ID)
				http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
				return
			}
			err := h.storage.SaveGauges(ctx, m.ID, *m.Value)
			if err != nil {
				h.logger.Errorw("ошибка сохранения gauge", "error", err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		case models.Counter:
			if m.Delta == nil {
				h.logger.Errorw("значение delta nil", "ID", m.ID)
				http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
				return
			}
			err := h.storage.SaveCounters(ctx, m.ID, *m.Delta)
			if err != nil {
				h.logger.Errorw("ошибка сохранения counter", "error", err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		default:
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		buf.Reset()
		if err := json.NewEncoder(buf).Encode(m); err != nil {
			h.logger.Errorw("ошибка сериализации", "error", err)
			return
		}
		w.Write(buf.Bytes())
	}
}

// ApiValueHandler возвращает обработчик для получения текущего значения метрики по её типу и названию.
func (h *MetricsHandler) ApiValueHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		var m models.Metrics
		buf := pool.Get().(*bytes.Buffer)
		buf.Reset()
		defer pool.Put(buf)
		if !strings.Contains(r.Header.Get("Content-Type"), "application/json") {
			h.logger.Errorw("неверный контент тип", "Content-Type", r.Header.Get("Content-Type"))
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		if _, err := buf.ReadFrom(r.Body); err != nil {
			h.logger.Errorw("ошибка чтения тела запроса", "error", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
			h.logger.Errorw("ошибка десериализации", "error", err)
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		if m.ID == "" {
			h.logger.Errorw("пустое имя метрики", "mType", m.MType)
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		switch m.MType {
		case models.Gauge:
			val, err := h.storage.GetGauges(ctx, m.ID)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					h.logger.Errorw("метрика gauge отсутствует в базе данных", "error", err)
					http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
					return
				}
				h.logger.Errorw("отсутствует подключение к БД", "error", err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			m.Value = &val
		case models.Counter:
			val, err := h.storage.GetCounters(ctx, m.ID)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					h.logger.Errorw("метрика counter отсутствует в базе данных", "error", err)
					http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
					return
				}
				h.logger.Errorw("отсутствует подключение к БД", "error", err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			m.Delta = &val
		default:
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		buf.Reset()
		if err := json.NewEncoder(buf).Encode(m); err != nil {
			h.logger.Errorw("ошибка сериализации", "error", err)
			return
		}
		w.Write(buf.Bytes())
	}
}

// ApiUpdatesHandler возвращает обработчик для пакетного сохранения метрик.
func (h *MetricsHandler) ApiUpdatesHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		var metrics []models.Metrics
		buf := pool.Get().(*bytes.Buffer)
		buf.Reset()
		defer pool.Put(buf)
		if !strings.Contains(r.Header.Get("Content-Type"), "application/json") {
			h.logger.Errorw("неверный контент тип", "Content-Type", r.Header.Get("Content-Type"))
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		if _, err := buf.ReadFrom(r.Body); err != nil {
			h.logger.Errorw("ошибка чтения тела запроса", "error", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		if err := json.Unmarshal(buf.Bytes(), &metrics); err != nil {
			h.logger.Errorw("ошибка десериализации", "error", err)
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		for _, m := range metrics {
			if m.MType != models.Gauge && m.MType != models.Counter {
				h.logger.Errorw("неверный тип метрики", "ID", m.ID)
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}
		}
		err := h.storage.SaveAll(ctx, metrics)
		if err != nil {
			h.logger.Errorw("ошибка сохранения", "error", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		buf.Reset()
		if err := json.NewEncoder(buf).Encode(metrics); err != nil {
			h.logger.Errorw("ошибка сериализации", "error", err)
			return
		}
		w.Write(buf.Bytes())
	}
}
