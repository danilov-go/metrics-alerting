package agent

import (
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"runtime"
	"time"
)

const defaultPort = "localhost:8080"

type Agent struct {
	Client http.Client
	Port   string
}

func New(port string) *Agent {
	if port == "" {
		port = defaultPort
	}
	return &Agent{
		Client: http.Client{
			Timeout: time.Second * 1,
		},
		Port: port,
	}
}

func (a *Agent) Run(pollCount *int64, step int) {
	var metricMap map[string]string
	metricMap = metricGet(pollCount)
	if *pollCount != 0 && int(*pollCount)%step == 0 {
		for name, val := range metricMap {
			metricType := "gauge"
			if name == "PollCount" {
				metricType = "counter"
			}
			url := fmt.Sprintf("http://%s/update/%s/%s/%s", a.Port, metricType, name, val)
			request, err := http.NewRequest(http.MethodPost, url, nil)
			if err != nil {
				fmt.Println(err)
				continue
			}
			request.Header.Set("Content-Type", "text/plain; charset=utf-8")
			response, err := a.Client.Do(request)
			if err != nil {
				fmt.Println(err)
				continue
			}
			_, err = io.Copy(io.Discard, response.Body)
			response.Body.Close()
			if err != nil {
				fmt.Println(err)
				continue
			}
			if response.StatusCode != http.StatusOK {
				fmt.Println(response.Status)
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
		"Alloc":         fmt.Sprintf("%f", float64(m.Alloc)),
		"BuckHashSys":   fmt.Sprintf("%f", float64(m.BuckHashSys)),
		"Frees":         fmt.Sprintf("%f", float64(m.Frees)),
		"GCCPUFraction": fmt.Sprintf("%f", float64(m.GCCPUFraction)),
		"GCSys":         fmt.Sprintf("%f", float64(m.GCSys)),
		"HeapAlloc":     fmt.Sprintf("%f", float64(m.HeapAlloc)),
		"HeapIdle":      fmt.Sprintf("%f", float64(m.HeapIdle)),
		"HeapInuse":     fmt.Sprintf("%f", float64(m.HeapInuse)),
		"HeapObjects":   fmt.Sprintf("%f", float64(m.HeapObjects)),
		"HeapReleased":  fmt.Sprintf("%f", float64(m.HeapReleased)),
		"HeapSys":       fmt.Sprintf("%f", float64(m.HeapSys)),
		"LastGC":        fmt.Sprintf("%f", float64(m.LastGC)),
		"Lookups":       fmt.Sprintf("%f", float64(m.Lookups)),
		"MCacheInuse":   fmt.Sprintf("%f", float64(m.MCacheInuse)),
		"MCacheSys":     fmt.Sprintf("%f", float64(m.MCacheSys)),
		"MSpanInuse":    fmt.Sprintf("%f", float64(m.MSpanInuse)),
		"MSpanSys":      fmt.Sprintf("%f", float64(m.MSpanSys)),
		"Mallocs":       fmt.Sprintf("%f", float64(m.Mallocs)),
		"NextGC":        fmt.Sprintf("%f", float64(m.NextGC)),
		"NumForcedGC":   fmt.Sprintf("%f", float64(m.NumForcedGC)),
		"NumGC":         fmt.Sprintf("%f", float64(m.NumGC)),
		"OtherSys":      fmt.Sprintf("%f", float64(m.OtherSys)),
		"PauseTotalNs":  fmt.Sprintf("%f", float64(m.PauseTotalNs)),
		"StackInuse":    fmt.Sprintf("%f", float64(m.StackInuse)),
		"StackSys":      fmt.Sprintf("%f", float64(m.StackSys)),
		"Sys":           fmt.Sprintf("%f", float64(m.Sys)),
		"TotalAlloc":    fmt.Sprintf("%f", float64(m.TotalAlloc)),
		"RandomValue":   fmt.Sprintf("%f", float64(randomValue)),
		"PollCount":     fmt.Sprintf("%d", *pollCount),
	}
	return metricMap
}
