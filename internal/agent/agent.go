package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"math/rand/v2"
	"net"
	"net/http"
	"runtime"
	"time"

	"github.com/danilov-go/metrics-alerting.git/internal/models"
	"github.com/go-resty/resty/v2"
)

type log interface {
	Errorw(msg string, keysAndValues ...any)
}

type Agent struct {
	Client *resty.Client
	logger log
}

func New(port string, l log) *Agent {
	client := resty.New()
	client.SetTimeout(time.Second * 1)
	client.SetBaseURL("http://" + port)
	return &Agent{
		Client: client,
		logger: l,
	}
}

func (a *Agent) Run(ctx context.Context, metrics []models.Metrics) {
	var buf bytes.Buffer
	wg := gzip.NewWriter(&buf)
	err := json.NewEncoder(wg).Encode(metrics)
	if err != nil {
		a.logger.Errorw("ошибка сжатия данных", "err", err)
		if errClose := wg.Close(); errClose != nil {
			a.logger.Errorw("ошибка закрытия gzip writer", "error", errClose)
		}
		return
	}
	err = wg.Close()
	if err != nil {
		a.logger.Errorw("ошибка закрытия gzip writer", "error", err)
		return
	}
	const maxRetries = 3
	duration := 1
	var response *resty.Response
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if err := ctx.Err(); err != nil {
			a.logger.Errorw("ошибка контекста", "error", err)
			return
		}
		response, err = a.Client.R().
			SetHeader("Content-Type", "application/json").
			SetHeader("Content-Encoding", "gzip").
			SetBody(buf.Bytes()).
			Post("/updates/")
		if err == nil {
			break
		}
		if attempt == maxRetries {
			a.logger.Errorw("попытки отправки исчерпаны", "maxRetries", maxRetries)
			return
		}
		var netErr net.Error
		if errors.As(err, &netErr) {
			timer := time.NewTimer(time.Duration(duration) * time.Second)
			select {
			case <-ctx.Done():
				timer.Stop()
				ctx.Err()
			case <-timer.C:
			}
			timer.Stop()
			duration += 2
			continue
		}
	}
	if response == nil {
		a.logger.Errorw("не удалось получить ответ от сервера", "error", err)
		return
	}
	if response.StatusCode() != http.StatusOK {
		a.logger.Errorw("статус запроса:", "status", response.StatusCode())
		return
	}
}

func (a *Agent) Get(pollCount int64) []models.Metrics {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	randomValue := rand.Float64()
	return []models.Metrics{
		{ID: "Alloc", MType: models.Gauge, Value: models.PointerFloat64(float64(m.Alloc))},
		{ID: "BuckHashSys", MType: models.Gauge, Value: models.PointerFloat64(float64(m.BuckHashSys))},
		{ID: "Frees", MType: models.Gauge, Value: models.PointerFloat64(float64(m.Frees))},
		{ID: "GCCPUFraction", MType: models.Gauge, Value: models.PointerFloat64(float64(m.GCCPUFraction))},
		{ID: "GCSys", MType: models.Gauge, Value: models.PointerFloat64(float64(m.GCSys))},
		{ID: "HeapAlloc", MType: models.Gauge, Value: models.PointerFloat64(float64(m.HeapAlloc))},
		{ID: "HeapIdle", MType: models.Gauge, Value: models.PointerFloat64(float64(m.HeapIdle))},
		{ID: "HeapInuse", MType: models.Gauge, Value: models.PointerFloat64(float64(m.HeapInuse))},
		{ID: "HeapObjects", MType: models.Gauge, Value: models.PointerFloat64(float64(m.HeapObjects))},
		{ID: "HeapReleased", MType: models.Gauge, Value: models.PointerFloat64(float64(m.HeapReleased))},
		{ID: "HeapSys", MType: models.Gauge, Value: models.PointerFloat64(float64(m.HeapSys))},
		{ID: "LastGC", MType: models.Gauge, Value: models.PointerFloat64(float64(m.LastGC))},
		{ID: "Lookups", MType: models.Gauge, Value: models.PointerFloat64(float64(m.Lookups))},
		{ID: "MCacheInuse", MType: models.Gauge, Value: models.PointerFloat64(float64(m.MCacheInuse))},
		{ID: "MCacheSys", MType: models.Gauge, Value: models.PointerFloat64(float64(m.MCacheSys))},
		{ID: "MSpanInuse", MType: models.Gauge, Value: models.PointerFloat64(float64(m.MSpanInuse))},
		{ID: "MSpanSys", MType: models.Gauge, Value: models.PointerFloat64(float64(m.MSpanSys))},
		{ID: "Mallocs", MType: models.Gauge, Value: models.PointerFloat64(float64(m.Mallocs))},
		{ID: "NextGC", MType: models.Gauge, Value: models.PointerFloat64(float64(m.NextGC))},
		{ID: "NumForcedGC", MType: models.Gauge, Value: models.PointerFloat64(float64(m.NumForcedGC))},
		{ID: "NumGC", MType: models.Gauge, Value: models.PointerFloat64(float64(m.NumGC))},
		{ID: "OtherSys", MType: models.Gauge, Value: models.PointerFloat64(float64(m.OtherSys))},
		{ID: "PauseTotalNs", MType: models.Gauge, Value: models.PointerFloat64(float64(m.PauseTotalNs))},
		{ID: "StackInuse", MType: models.Gauge, Value: models.PointerFloat64(float64(m.StackInuse))},
		{ID: "StackSys", MType: models.Gauge, Value: models.PointerFloat64(float64(m.StackSys))},
		{ID: "Sys", MType: models.Gauge, Value: models.PointerFloat64(float64(m.Sys))},
		{ID: "TotalAlloc", MType: models.Gauge, Value: models.PointerFloat64(float64(m.TotalAlloc))},
		{ID: "RandomValue", MType: models.Gauge, Value: models.PointerFloat64(float64(randomValue))},
		{ID: "PollCount", MType: models.Counter, Delta: models.PointerInt64(pollCount)},
	}
}
