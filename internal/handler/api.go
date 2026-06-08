package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"

	"net/http"

	"github.com/danilov-go/metrics-alerting.git/internal/models"
)

func ApiUpdateHandler(ctx context.Context, s Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var m models.Metrics
		var buf bytes.Buffer
		if !strings.Contains(r.Header.Get("Content-Type"), "application/json") {
			http.Error(w, "не верный контент тип", http.StatusBadRequest)
			return
		}
		_, err := buf.ReadFrom(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err = json.Unmarshal(buf.Bytes(), &m); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if m.ID == "" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		switch m.MType {
		case models.Gauge:
			if m.Value == nil {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			s.SaveGauges(ctx, m.ID, *m.Value)
		case models.Counter:
			if m.Delta == nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			s.SaveCounters(ctx, m.ID, *m.Delta)
		default:
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		resp, err := json.Marshal(m)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(resp)
	}
}

func ApiValueHandler(ctx context.Context, s Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var m models.Metrics
		var buf bytes.Buffer
		if !strings.Contains(r.Header.Get("Content-Type"), "application/json") {
			http.Error(w, "не верный контент тип", http.StatusBadRequest)
			return
		}
		_, err := buf.ReadFrom(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err = json.Unmarshal(buf.Bytes(), &m); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if m.ID == "" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		switch m.MType {
		case models.Gauge:
			val, err := s.GetGauges(ctx, m.ID)
			if err != nil {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			m.Value = &val
		case models.Counter:
			val, err := s.GetCounters(ctx, m.ID)
			if err != nil {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			m.Delta = &val
		default:
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		resp, err := json.Marshal(m)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(resp)
	}
}

func ApiUpdatesHandler(ctx context.Context, s Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var metrics []models.Metrics
		var buf bytes.Buffer
		if !strings.Contains(r.Header.Get("Content-Type"), "application/json") {
			http.Error(w, "не верный контент тип", http.StatusBadRequest)
			return
		}
		_, err := buf.ReadFrom(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err = json.Unmarshal(buf.Bytes(), &metrics); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		err = s.SaveAll(ctx, metrics)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		resp, err := json.Marshal(metrics)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(resp)
	}
}
