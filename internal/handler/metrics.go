package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/danilov-go/metrics-alerting.git/internal/models"
)

type Storage interface {
	SaveGauges(name string, value float64)
	SaveCounters(name string, value int64)
	GetGauges(name string) (float64, bool)
	GetCounters(name string) (int64, bool)
	GetAllGauges() map[string]float64
	GetAllCounters() map[string]int64
}

func PostMetricsHandler(s Storage) http.HandlerFunc {
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

func GetMetricHandler(s Storage) http.HandlerFunc {
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

func ExposeMetricsHandler(s Storage) http.HandlerFunc {
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
			fmt.Fprintf(w, "%v : %v<br>\n", name, val)
		}
		for name, val := range mapCounter {
			fmt.Fprintf(w, "%v : %v<br>\n", name, val)
		}
	}
}
