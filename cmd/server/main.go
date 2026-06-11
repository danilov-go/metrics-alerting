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
	var storage handler.Storage
	configs := config.ConfigServer{
		Net: config.NetAddress{
			Host: "localhost",
			Port: 8080,
		},
		StoreIntrval:    300,
		FileStoragePath: "metricStorage.txt",
		Restore:         false,
		DatabaseDsn:     "host=localhost user=metrics password=123 dbname=metrics sslmode=disable",
	}
	if err := logger.Initialize("info"); err != nil {
		panic(err)
	}
	configs.Get()
	cfg := repository.ConfigFile{
		Path:     configs.FileStoragePath,
		Interval: time.Duration(configs.StoreIntrval) * time.Second,
		Restore:  configs.Restore,
	}
	pg, err := db.InitDB(configs.DatabaseDsn)
	if err != nil {
		logger.Log.Info("не удалось подключится к базе данных", zap.Error(err))
	}
	switch {
	case configs.ValidDB == true && err == nil:
		storage = handler.NewErrorMiddleware(pg)
	case configs.ValidFile == true:
		storage = repository.InitMemStorage(cfg, logger.Log.Sugar())
	default:
		storage = repository.InitMemStorage(cfg, logger.Log.Sugar())
	}
	h := handler.NewMetricsHandler(storage, logger.Log.Sugar())
	r := chi.NewRouter()
	r.Use(handler.RequestLogger(logger.Log))
	r.Use(handler.GzipMiddleware)
	r.Post("/update/{mType}/{mName}/{mVal}", h.PostMetricsHandler())
	r.Get("/value/{mType}/{mName}", h.GetMetricHandler())
	r.Post("/updates", h.ApiUpdatesHandler())
	r.Post("/updates/", h.ApiUpdatesHandler())
	r.Post("/update", h.ApiUpdateHandler())
	r.Post("/update/", h.ApiUpdateHandler())
	r.Post("/value", h.ApiValueHandler())
	r.Post("/value/", h.ApiValueHandler())
	r.Get("/ping", h.PingHandler())
	r.Get("/", h.ExposeMetricsHandler())
	serv := server.New(configs.Net.String(), logger.Log.Sugar(), r)
	if err := serv.Run(); err != nil {
		panic(err)
	}
}
