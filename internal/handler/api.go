package handler

import (
	"bytes"
	"encoding/json"

	"net/http"

	"github.com/danilov-go/metrics-alerting.git/internal/models"
)

func ApiUpdateHandler(s Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var m models.Metrics
		var buf bytes.Buffer
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
			s.SaveGauges(m.ID, *m.Value)
		case models.Counter:
			s.SaveCounters(m.ID, *m.Delta)
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

func ApiValueHandler(s Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var m models.Metrics
		var buf bytes.Buffer
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
			val, valid := s.GetGauges(m.ID)
			if !valid {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			m.Value = &val
		case models.Counter:
			val, valid := s.GetCounters(m.ID)
			if !valid {
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
