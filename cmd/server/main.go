package main

import (
	"time"

	"github.com/danilov-go/metrics-alerting.git/internal/config"
	"github.com/danilov-go/metrics-alerting.git/internal/config/db"
	"github.com/danilov-go/metrics-alerting.git/internal/handler"
	"github.com/danilov-go/metrics-alerting.git/internal/logger"
	"github.com/danilov-go/metrics-alerting.git/internal/repository"
	"github.com/danilov-go/metrics-alerting.git/internal/server"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func main() {
	configs := config.ConfigServer{
		Net: config.NetAddress{
			Host: "localhost",
			Port: 8080,
		},
		StoreIntrval:    300,
		FileStoragePath: "metricStorage.txt",
		Restore:         true,
		DatabaseDsn:     "host=localhost user=video password=123 dbname=video sslmode=disable",
	}
	configs.Get()
	if err := logger.Initialize("info"); err != nil {
		panic(err)
	}
	metricsStorage := &repository.MemStorage{
		Gauges:   make(map[string]float64),
		Counters: make(map[string]int64),
	}
	err := db.InitDB(configs.DatabaseDsn)
	if err != nil {
		panic(err)
	}
	defer db.DB.Close()
	if configs.Restore {
		err := metricsStorage.LoadFile(configs.FileStoragePath)
		if err != nil {
			logger.Log.Info("не удалось восстановить метрики из файла", zap.Error(err))
		}
	}
	if configs.StoreIntrval > 0 {
		storeIntrval := time.Duration(configs.StoreIntrval) * time.Second
		metricsStorage.Run(configs.FileStoragePath, storeIntrval)
	}
	r := chi.NewRouter()
	r.Use(handler.RequestLogger(logger.Log))
	r.Use(handler.GzipMiddleware)
	r.Post("/update/{mType}/{mName}/{mVal}", handler.PostMetricsHandler(metricsStorage))
	r.Get("/value/{mType}/{mName}", handler.GetMetricHandler(metricsStorage))
	r.Post("/update", handler.ApiUpdateHandler(metricsStorage, configs))
	r.Post("/update/", handler.ApiUpdateHandler(metricsStorage, configs))
	r.Post("/value", handler.ApiValueHandler(metricsStorage))
	r.Post("/value/", handler.ApiValueHandler(metricsStorage))
	r.Get("/ping", handler.PingHandler(db.DB))
	r.Get("/", handler.ExposeMetricsHandler(metricsStorage))
	serv := server.New(configs.Net.String(), logger.Log.Sugar(), r)
	if err := serv.Run(); err != nil {
		panic(err)
	}
}
