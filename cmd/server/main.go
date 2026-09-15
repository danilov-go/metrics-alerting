package main

import (
	"context"
	_ "net/http/pprof"
	"os/signal"
	"syscall"

	"github.com/danilov-go/metrics-alerting.git/internal/config"
	"github.com/danilov-go/metrics-alerting.git/internal/logger"
	"github.com/danilov-go/metrics-alerting.git/internal/server"
	"golang.org/x/sync/errgroup"
)

var (
	buildVersion string = "N/A"
	buildDate    string = "N/A"
	buildCommit  string = "N/A"
)

func main() {
	config.PrintBuild(buildVersion, buildDate, buildCommit)
	configs := config.ConfigServer{
		Net: config.NetAddress{
			Host: "localhost",
			Port: 8080,
		},
		StoreIntrval:    300,
		FileStoragePath: "metricStorage.txt",
		Restore:         false,
		DatabaseDSN:     "",
		Key:             "",
		AuditFile:       "",
		AuditURL:        "",
		TrustedSubnet:   "",
		RetryDuration:   1,
		RetryInterval:   2,
	}
	if err := logger.Initialize("info"); err != nil {
		panic(err)
	}
	if err := configs.Get(); err != nil {
		logger.Log.Sugar().Fatal("ошибка загрузки конфигурации сервера")
	}
	storage := initStorage(configs)
	defer func() {
		if err := storage.Close(); err != nil {
			logger.Log.Sugar().Errorw("ошибка закрытия хралища", "error", err)
		}
	}()
	clientAudit := initClient(configs)
	event, fileSub, urlSub := initAudit(configs, clientAudit)
	if fileSub != nil {
		defer fileSub.Close()
	}
	if urlSub != nil {
		defer urlSub.Close()
	}
	r := initRouter(configs, storage, event, fileSub, urlSub)
	runProffServer()
	serv := server.New(configs.Net.String(), logger.Log.Sugar(), r)
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer stop()
	g, gCtx := errgroup.WithContext(ctx)
	g.Go(func() error {
		return serv.Run()
	})
	g.Go(func() error {
		<-gCtx.Done()
		return serv.Stop()
	})
	if err := g.Wait(); err != nil {
		logger.Log.Sugar().Fatal("сервер аварийно завершил работу", "err", err)
	}
}
