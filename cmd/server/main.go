package main

import (
	"context"
	"crypto/rsa"
	_ "net/http/pprof"
	"os"
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
	config.PrintBuild(os.Stdout, buildVersion, buildDate, buildCommit)
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
		GrpcAddress:     "",
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
	ipNet, err := config.ParseTrustedSubnet(configs.TrustedSubnet)
	if err != nil {
		logger.Log.Sugar().Fatal("ошибка парсинга TrustedSubnet")
	}
	var privatKey *rsa.PrivateKey
	if configs.CryptoKey != "" {
		privatKey, err = configs.GetKey()
		if err != nil {
			logger.Log.Sugar().Fatal("ошибка загрузки или парсинга приватного ключа сервера")
		}
	}
	grpcServ, err := initGRPC(configs, ipNet, storage, logger.Log, privatKey)
	if err != nil {
		logger.Log.Sugar().Fatalw("ошибка инициализации gRPC сервера", "error", err)
	}
	r := initRouter(configs, storage, event, fileSub, urlSub, ipNet, privatKey)
	runServerPprof()
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
	if grpcServ != nil {
		g.Go(func() error {
			return grpcServ.Run()
		})
		g.Go(func() error {
			<-gCtx.Done()
			return grpcServ.Stop()
		})
	}
	if err := g.Wait(); err != nil {
		logger.Log.Sugar().Fatal("сервер аварийно завершил работу", "err", err)
	}
}
