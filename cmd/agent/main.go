package main

import (
	"context"

	"github.com/danilov-go/metrics-alerting.git/internal/agent"
	"github.com/danilov-go/metrics-alerting.git/internal/config"
	"github.com/danilov-go/metrics-alerting.git/internal/logger"
)

func main() {
	configs := &config.ConfigAgent{
		Net: config.NetAddress{
			Host: "localhost",
			Port: 8080,
		},
		PollInterval:   2,
		ReportInterval: 10,
		Key:            "",
		RateLimit:      3,
	}
	configs.Get()
	if err := logger.Initialize("info"); err != nil {
		panic(err)
	}
	logger.Log.Sugar().Info("Key", configs.Key)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	client := agent.New(*configs, logger.Log.Sugar())
	client.Run(ctx)
}
