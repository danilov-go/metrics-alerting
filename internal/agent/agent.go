// Package agent собирает метрики runtime/gopsutil,
// и с заданной периодичностью отправляет на сервер.
package agent

import (
	"context"
	"crypto/rsa"
	"sync"
	"sync/atomic"
	"time"

	"github.com/danilov-go/metrics-alerting.git/internal/config"
	"github.com/danilov-go/metrics-alerting.git/internal/models"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// MetricSender определяет общий интерфейс для отправки метрик.
type MetricSender interface {
	Send(ctx context.Context, metrics []models.Metrics, publicKey *rsa.PublicKey)
}

type log interface {
	Errorw(msg string, keysAndValues ...any)
	Fatalw(msg string, keysAndValues ...any)
}

// Agent собирает метрики runtime/gopsutil и передает их на сервер.
type Agent struct {
	Sender         MetricSender
	logger         log
	pollInterval   int
	reportInterval int
	rateLimit      int
	pollCount      atomic.Int64
}

// New создает новый экземпляр Agent.
func New(cfg config.ConfigAgent, l log) (*Agent, error) {
	agent := &Agent{
		logger:         l,
		pollInterval:   int(cfg.PollInterval),
		reportInterval: int(cfg.ReportInterval),
		rateLimit:      cfg.RateLimit,
	}
	if cfg.GrpcAddress != "" {
		conn, err := grpc.NewClient(cfg.GrpcAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return nil, err
		}
		host, err := config.GetHost(cfg.GrpcAddress)
		if err != nil {
			l.Fatalw("ошибка получения host агента", "error", err)
		}
		grpcSender, err := NewGRPCSender(conn, host, l)
		if err != nil {
			l.Fatalw("не удалось инициализировать gRPC клиент", "error", err)
		}
		agent.Sender = grpcSender
		return agent, nil
	}
	host, err := config.GetHost(cfg.Net.String())
	if err != nil {
		l.Fatalw("ошибка получения host агента", "error", err)
	}
	agent.Sender = NewHTTPSender(cfg.Net.String(), cfg.Key, host, l)
	return agent, nil
}

// Run запускает процесс сбора и отправки метрик на сервер.
func (a *Agent) Run(ctx context.Context, publicKey *rsa.PublicKey) {
	var wg sync.WaitGroup
	gopsutilChan := a.getGopsutil(ctx, &wg)
	runtimeChan := a.getRuntime(ctx, &wg)
	metricChan := a.merge(ctx, &wg, gopsutilChan, runtimeChan)
	a.worker(ctx, &wg, metricChan, publicKey)
	<-ctx.Done()
	wg.Wait()
}

func (a *Agent) worker(ctx context.Context, wg *sync.WaitGroup, metricsChan chan []models.Metrics, publicKey *rsa.PublicKey) {
	for i := 0; i < a.rateLimit; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for ch := range metricsChan {
				if len(ch) == 0 {
					continue
				}
				a.Sender.Send(ctx, ch, publicKey)
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
				if len(metrics) > 0 {
					metricsChan <- metrics
				}
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
