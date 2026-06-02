package main

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/danilov-go/metrics-alerting.git/internal/agent"
	"github.com/danilov-go/metrics-alerting.git/internal/config"
	"github.com/danilov-go/metrics-alerting.git/internal/logger"
	"github.com/danilov-go/metrics-alerting.git/internal/models"
)

func main() {
	configs := &config.ConfigAgent{
		Net: config.NetAddress{
			Host: "localhost",
			Port: 8080,
		},
		PollInterval:   2,
		ReportInterval: 10,
	}
	configs.Get()
	if err := logger.Initialize("info"); err != nil {
		panic(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	pollInterval := time.Duration(configs.PollInterval) * time.Second
	reportInterval := time.Duration(configs.ReportInterval) * time.Second
	tikerPoll := time.NewTicker(pollInterval)
	tikerReport := time.NewTicker(reportInterval)
	client := agent.New(configs.Net.String(), logger.Log.Sugar())
	var pollCount atomic.Int64
	var metrics []models.Metrics
	for {
		select {
		case <-ctx.Done():
			logger.Log.Info("получена команда на завершение")
			return
		case <-tikerPoll.C:
			pollCount.Add(1)
			metrics = client.Get(pollCount.Load())
		case <-tikerReport.C:
			client.Run(metrics)
		}
	}
}
