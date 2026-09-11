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
	config.PrintBuild(buildVersion, buildDate, buildCommit)
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
	if err := configs.Get(); err != nil {
		panic(err)
	}
	if err := logger.Initialize("info"); err != nil {
		panic(err)
	}
	var publicKey *rsa.PublicKey
	var errC error
	if configs.CryptoKey != "" {
		if publicKey, errC = configs.GetKey(); errC != nil {
			panic(errC)
		}
	}
	logger.Log.Sugar().Info("Key", configs.Key)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		signalChan := make(chan os.Signal, 1)
		signal.Notify(signalChan, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
		<-signalChan
		cancel()
	}()
	client := agent.New(*configs, logger.Log.Sugar())
	client.Run(ctx, publicKey)
}
