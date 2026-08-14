package agent

import (
	"fmt"
	"math/rand/v2"
	"runtime"

	"github.com/danilov-go/metrics-alerting.git/internal/models"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
)

func getRuntime(pollCount int64) []models.Metrics {
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

func getGopsutil() ([]models.Metrics, error) {
	var metrics []models.Metrics
	v, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}
	metrics = append(metrics, models.Metrics{ID: "TotalMemory", MType: models.Gauge, Value: models.PointerFloat64(float64(v.Total))})
	metrics = append(metrics, models.Metrics{ID: "FreeMemory", MType: models.Gauge, Value: models.PointerFloat64(float64(v.Free))})
	p, err := cpu.Percent(0, true)
	if err != nil {
		return nil, err
	}
	for i, c := range p {
		metrics = append(metrics, models.Metrics{ID: fmt.Sprintf("CPUutilization%d", i+1), MType: models.Gauge, Value: models.PointerFloat64(float64(c))})
	}
	return metrics, nil
}
