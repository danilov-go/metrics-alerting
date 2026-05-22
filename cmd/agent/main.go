package main

import (
	"time"

	"github.com/danilov-go/metrics-alerting.git/internal/agent"
	"github.com/danilov-go/metrics-alerting.git/internal/config"
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
	pollInterval := time.Duration(configs.PollInterval) * time.Second
	reportInterval := time.Duration(configs.ReportInterval) * time.Second
	step := int64(reportInterval / pollInterval)
	client := agent.New(configs.Net.String())
	var pollCount int64 = 0
	for {
		client.Run(&pollCount, step)
		time.Sleep(pollInterval)
	}
}
