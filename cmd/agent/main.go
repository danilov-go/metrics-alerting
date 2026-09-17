package main

import (
	"context"
	"crypto/rsa"
	"os"
	"os/signal"
	"syscall"

	"github.com/danilov-go/metrics-alerting.git/internal/agent"
	"github.com/danilov-go/metrics-alerting.git/internal/config"
	"github.com/danilov-go/metrics-alerting.git/internal/logger"
)

var (
	buildVersion string = "N/A"
	buildDate    string = "N/A"
	buildCommit  string = "N/A"
)

func main() {
	config.PrintBuild(os.Stdout, buildVersion, buildDate, buildCommit)
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
	if err := logger.Initialize("info"); err != nil {
		panic(err)
	}
	if err := configs.Get(); err != nil {
		logger.Log.Sugar().Fatal("ошибка загрузки конфигурации агента")
	}
	var publicKey *rsa.PublicKey
	var errC error
	if configs.CryptoKey != "" {
		if publicKey, errC = configs.GetKey(); errC != nil {
			logger.Log.Sugar().Fatal("ошибка загрузки или парсинга публичного ключа агента")
		}
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer stop()
	client := agent.New(*configs, logger.Log.Sugar())
	client.Run(ctx, publicKey)
}
