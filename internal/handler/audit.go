package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/danilov-go/metrics-alerting.git/internal/audit"
	"github.com/go-chi/chi/v5"
)

type metricName struct {
	ID string `json:"id"`
}

type auditWriter struct {
	w      http.ResponseWriter
	status int
}

func (a *auditWriter) Header() http.Header {
	return a.w.Header()
}

func (a *auditWriter) WriteHeader(statusCode int) {
	a.status = statusCode
	a.w.WriteHeader(statusCode)
}

func (a *auditWriter) Write(b []byte) (int, error) {
	if a.status == 0 {
		a.status = http.StatusOK
	}
	return a.w.Write(b)
}

// AuditMiddleware отправляет успешные обновления метрик в адуит.
func AuditMiddleware(event audit.Publisher) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var metrics []string
			if r.Body != nil {
				body, err := io.ReadAll(r.Body)
				if err == nil {
					r.Body.Close()
					r.Body = io.NopCloser(bytes.NewReader(body))
					var metric []metricName
					if err := json.Unmarshal(body, &metric); err == nil {
						for _, m := range metric {
							if m.ID != "" {
								metrics = append(metrics, m.ID)
							}
						}
					} else {
						var m metricName
						if err := json.Unmarshal(body, &m); err == nil && m.ID != "" {
							metrics = append(metrics, m.ID)
						}
					}
				}
			}
			a := &auditWriter{w: w}
			next.ServeHTTP(a, r)
			if a.status == 0 {
				a.status = http.StatusOK
			}
			if a.status >= 400 {
				return
			}
			if mName := chi.URLParam(r, "mName"); mName != "" {
				metrics = append(metrics, mName)
			}
			if len(metrics) > 0 {
				ip, _, err := net.SplitHostPort(r.RemoteAddr)
				if err != nil {
					ip = r.RemoteAddr
				}
				auditLog := audit.Audit{
					TS:        time.Now().Unix(),
					Metrics:   metrics,
					IPAddress: ip,
				}
				event.Notify(auditLog)
			}
		})
	}
}
