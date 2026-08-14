package agent

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/danilov-go/metrics-alerting.git/internal/config"
	"github.com/danilov-go/metrics-alerting.git/internal/models"
	"github.com/go-resty/resty/v2"
)

type log interface {
	Errorw(msg string, keysAndValues ...any)
}

type Agent struct {
	Client         *resty.Client
	logger         log
	key            string
	pollInterval   int
	reportInterval int
	rateLimit      int
	pollCount      atomic.Int64
}

func New(cfg config.ConfigAgent, l log) *Agent {
	client := resty.New()
	client.SetTimeout(time.Second * 1)
	client.SetBaseURL("http://" + cfg.Net.String())
	fmt.Println(cfg.Net.String())
	return &Agent{
		Client:         client,
		logger:         l,
		key:            cfg.Key,
		pollInterval:   cfg.PollInterval,
		reportInterval: cfg.ReportInterval,
		rateLimit:      cfg.RateLimit,
	}
}

func (a *Agent) Run(ctx context.Context) {
	var wg sync.WaitGroup
	gopsutilChan := a.getGopsutil(ctx, &wg)
	runtimeChan := a.getRuntime(ctx, &wg)
	metricChan := a.merge(ctx, &wg, gopsutilChan, runtimeChan)
	a.worker(ctx, &wg, metricChan)
	<-ctx.Done()
	wg.Wait()
}

func (a *Agent) worker(ctx context.Context, wg *sync.WaitGroup, metricsChan chan []models.Metrics) {
	for i := 0; i < a.rateLimit; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for ch := range metricsChan {
				if len(ch) == 0 {
					continue
				}
				a.send(ctx, ch)
			}
		}()
	}
}

func (a *Agent) merge(ctx context.Context, wg *sync.WaitGroup, gopsutilChan, runtimeChan chan []models.Metrics) chan []models.Metrics {
	metricsChan := make(chan []models.Metrics, a.rateLimit)
	wg.Add(1)
	go func() {
		defer wg.Done()
		tikerReport := time.NewTicker(time.Duration(a.reportInterval) * time.Second)
		defer tikerReport.Stop()
		defer close(metricsChan)
		var metrics []models.Metrics
		for {
			select {
			case <-ctx.Done():
				return
			case metric, ok := <-gopsutilChan:
				if !ok {
					gopsutilChan = nil
					continue
				}
				metrics = append(metrics, metric...)
			case metric, ok := <-runtimeChan:
				if !ok {
					runtimeChan = nil
					continue
				}
				metrics = append(metrics, metric...)
			case <-tikerReport.C:
				if len(metrics) == 0 {
					continue
				}
				metricsChan <- metrics
				metrics = nil
			}
		}
	}()
	return metricsChan
}

func (a *Agent) getGopsutil(ctx context.Context, wg *sync.WaitGroup) chan []models.Metrics {
	tikerPoll := time.NewTicker(time.Duration(a.pollInterval) * time.Second)
	gopsutilChan := make(chan []models.Metrics, a.rateLimit)
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer tikerPoll.Stop()
		defer close(gopsutilChan)
		for {
			select {
			case <-ctx.Done():
				return
			case <-tikerPoll.C:
				metricsGopsutil, err := getGopsutil()
				if err != nil {
					a.logger.Errorw("ошибка сбора метрик", "error", err)
				}
				if len(metricsGopsutil) > 0 {
					gopsutilChan <- metricsGopsutil
				}
			}
		}
	}()
	return gopsutilChan
}

func (a *Agent) getRuntime(ctx context.Context, wg *sync.WaitGroup) chan []models.Metrics {
	tikerPoll := time.NewTicker(time.Duration(a.pollInterval) * time.Second)
	runtimeChan := make(chan []models.Metrics, a.rateLimit)
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer tikerPoll.Stop()
		defer close(runtimeChan)
		for {
			select {
			case <-ctx.Done():
				return
			case <-tikerPoll.C:
				a.pollCount.Add(1)
				metricsRuntime := getRuntime(a.pollCount.Load())
				if len(metricsRuntime) > 0 {
					runtimeChan <- metricsRuntime
				}
			}
		}
	}()
	return runtimeChan
}
