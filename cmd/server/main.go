package main

import (
	"context"
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
	ctx := context.Background()
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
	var dbPing *sql.DB
	Pg, err := db.InitDB(configs.DatabaseDsn, logger.Log.Sugar())
	if err != nil {
		logger.Log.Info("не удалось подключится к базе данных", zap.Error(err))
	}
	if Pg != nil {
		dbPing = Pg.DB
		defer Pg.DB.Close()
	}
	switch {
	case configs.ValidDB == true && err == nil:
		storage = handler.NewErrorMiddleware(Pg)
	case configs.ValidFile == true:
		storage = repository.InitMemStorage(logger.Log.Sugar(), cfg)
	default:
		storage = repository.InitMemStorage(logger.Log.Sugar(), cfg)
	}
	r := chi.NewRouter()
	r.Use(handler.RequestLogger(logger.Log))
	r.Use(handler.GzipMiddleware)
	r.Post("/update/{mType}/{mName}/{mVal}", handler.PostMetricsHandler(ctx, storage))
	r.Get("/value/{mType}/{mName}", handler.GetMetricHandler(ctx, storage))
	r.Post("/updates", handler.ApiUpdatesHandler(ctx, storage))
	r.Post("/updates/", handler.ApiUpdatesHandler(ctx, storage))
	r.Post("/update", handler.ApiUpdateHandler(ctx, storage))
	r.Post("/update/", handler.ApiUpdateHandler(ctx, storage))
	r.Post("/value", handler.ApiValueHandler(ctx, storage))
	r.Post("/value/", handler.ApiValueHandler(ctx, storage))
	r.Get("/ping", handler.PingHandler(dbPing))
	r.Get("/", handler.ExposeMetricsHandler(ctx, storage))
	serv := server.New(configs.Net.String(), logger.Log.Sugar(), r)
	if err := serv.Run(); err != nil {
		panic(err)
	}
}
