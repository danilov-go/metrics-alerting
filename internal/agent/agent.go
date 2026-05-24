package agent

import (
	"fmt"
	"math/rand/v2"
	"net/http"
	"runtime"
	"time"

	"github.com/go-resty/resty/v2"
)

type log interface {
	Error(args ...any)
}

type Agent struct {
	Client *resty.Client
	Logger log
}

func New(port string, l log) *Agent {
	client := resty.New()
	client.SetTimeout(time.Second * 1)
	client.SetBaseURL("http://" + port)
	return &Agent{
		Client: client,
		Logger: l,
	}
}

func (a *Agent) Run(pollCount *int64, step int64) {
	var metricMap map[string]string
	metricMap = metricGet(pollCount)
	if *pollCount != 0 && *pollCount%step == 0 {
		for name, val := range metricMap {
			metricType := "gauge"
			if name == "PollCount" {
				metricType = "counter"
			}
			response, err := a.Client.R().
				SetHeader("Content-Type", "text/plain; charset=utf-8").
				SetPathParams(map[string]string{
					"mType": metricType,
					"mName": name,
					"mVal":  val,
				}).
				Post("/update/{mType}/{mName}/{mVal}")
			if err != nil {
				a.Logger.Error(err.Error())
				continue
			}
			if response.StatusCode() != http.StatusOK {
				a.Logger.Error(response.Status())
				continue
			}
		}
	}
}

func metricGet(pollCount *int64) map[string]string {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	randomValue := rand.Float64()
	*pollCount++
	metricMap := map[string]string{
		"Alloc":         fmt.Sprintf("%v", float64(m.Alloc)),
		"BuckHashSys":   fmt.Sprintf("%v", float64(m.BuckHashSys)),
		"Frees":         fmt.Sprintf("%v", float64(m.Frees)),
		"GCCPUFraction": fmt.Sprintf("%v", float64(m.GCCPUFraction)),
		"GCSys":         fmt.Sprintf("%v", float64(m.GCSys)),
		"HeapAlloc":     fmt.Sprintf("%v", float64(m.HeapAlloc)),
		"HeapIdle":      fmt.Sprintf("%v", float64(m.HeapIdle)),
		"HeapInuse":     fmt.Sprintf("%v", float64(m.HeapInuse)),
		"HeapObjects":   fmt.Sprintf("%v", float64(m.HeapObjects)),
		"HeapReleased":  fmt.Sprintf("%v", float64(m.HeapReleased)),
		"HeapSys":       fmt.Sprintf("%v", float64(m.HeapSys)),
		"LastGC":        fmt.Sprintf("%v", float64(m.LastGC)),
		"Lookups":       fmt.Sprintf("%v", float64(m.Lookups)),
		"MCacheInuse":   fmt.Sprintf("%v", float64(m.MCacheInuse)),
		"MCacheSys":     fmt.Sprintf("%v", float64(m.MCacheSys)),
		"MSpanInuse":    fmt.Sprintf("%v", float64(m.MSpanInuse)),
		"MSpanSys":      fmt.Sprintf("%v", float64(m.MSpanSys)),
		"Mallocs":       fmt.Sprintf("%v", float64(m.Mallocs)),
		"NextGC":        fmt.Sprintf("%v", float64(m.NextGC)),
		"NumForcedGC":   fmt.Sprintf("%v", float64(m.NumForcedGC)),
		"NumGC":         fmt.Sprintf("%v", float64(m.NumGC)),
		"OtherSys":      fmt.Sprintf("%v", float64(m.OtherSys)),
		"PauseTotalNs":  fmt.Sprintf("%v", float64(m.PauseTotalNs)),
		"StackInuse":    fmt.Sprintf("%v", float64(m.StackInuse)),
		"StackSys":      fmt.Sprintf("%v", float64(m.StackSys)),
		"Sys":           fmt.Sprintf("%v", float64(m.Sys)),
		"TotalAlloc":    fmt.Sprintf("%v", float64(m.TotalAlloc)),
		"RandomValue":   fmt.Sprintf("%v", float64(randomValue)),
		"PollCount":     fmt.Sprintf("%v", *pollCount),
	}
	return metricMap
}
