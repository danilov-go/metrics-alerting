package main

import (
	"database/sql"
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
	metricsStorage := &repository.MemStorage{
		Gauges:   make(map[string]float64),
		Counters: make(map[string]int64),
	}
	configs.Get()
	var sqlPg *sql.DB
	Pg, err := db.InitDB(configs.DatabaseDsn)
	if err != nil {
		logger.Log.Info("не удалось подключится к базе данных", zap.Error(err))
	} else {
		sqlPg = Pg.DB
	}
	if sqlPg != nil {
		defer sqlPg.Close()
	}
	switch {
	case configs.ValidDB == true && err == nil:
		storage = Pg
	case configs.ValidFile == true:
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
		storage = metricsStorage
	default:
		storage = metricsStorage
	}
	r := chi.NewRouter()
	r.Use(handler.RequestLogger(logger.Log))
	r.Use(handler.GzipMiddleware)
	r.Post("/update/{mType}/{mName}/{mVal}", handler.PostMetricsHandler(storage))
	r.Get("/value/{mType}/{mName}", handler.GetMetricHandler(storage))
	r.Post("/update", handler.ApiUpdateHandler(storage, configs))
	r.Post("/update/", handler.ApiUpdateHandler(storage, configs))
	r.Post("/value", handler.ApiValueHandler(storage))
	r.Post("/value/", handler.ApiValueHandler(storage))
	r.Get("/ping", handler.PingHandler(sqlPg))
	r.Get("/", handler.ExposeMetricsHandler(storage))
	serv := server.New(configs.Net.String(), logger.Log.Sugar(), r)
	if err := serv.Run(); err != nil {
		panic(err)
	}
}
