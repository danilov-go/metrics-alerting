package main

import (
	"net/http"
	_ "net/http/pprof"
	"time"

	"github.com/danilov-go/metrics-alerting.git/internal/audit"
	"github.com/danilov-go/metrics-alerting.git/internal/config"
	"github.com/danilov-go/metrics-alerting.git/internal/config/db"
	"github.com/danilov-go/metrics-alerting.git/internal/handler"
	"github.com/danilov-go/metrics-alerting.git/internal/logger"
	"github.com/danilov-go/metrics-alerting.git/internal/repository"
	"github.com/danilov-go/metrics-alerting.git/internal/server"
	"github.com/go-chi/chi/v5"
	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

var (
	buildVersion string = "N/A"
	buildDate    string = "N/A"
	buildCommit  string = "N/A"
)

func main() {
	config.PrintBuild(buildVersion, buildDate, buildCommit)
	var storage handler.Storage
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
		RetryDuration:   1,
		RetryInterval:   2,
	}
	if err := logger.Initialize("info"); err != nil {
		panic(err)
	}
	configs.Get()
	if configs.ValidDB && configs.DatabaseDSN == "" {
		panic("DatabaseDSN передан, но является пустым")
	}
	cfg := repository.ConfigFile{
		Path:     configs.FileStoragePath,
		Interval: time.Duration(configs.StoreIntrval) * time.Second,
		Restore:  configs.Restore,
	}
	var pg handler.Storage
	var dbErr error
	if configs.ValidDB {
		pg, dbErr = db.InitDB(configs.DatabaseDSN)
		if dbErr != nil {
			logger.Log.Info("не удалось подключится к базе данных, переключаемся на memstorage", zap.Error(dbErr))
		}
	}
	if configs.ValidDB && dbErr == nil && pg != nil {
		duration := time.Duration(configs.RetryDuration) * time.Second
		interval := time.Duration(configs.RetryInterval) * time.Second
		storage = handler.NewErrorMiddleware(pg, duration, interval)
	} else {
		storage = repository.InitMemStorage(cfg, logger.Log.Sugar())
	}
	logger.Log.Sugar().Info("Key", configs.Key)
	client := resty.New().
		SetTimeout(5 * time.Second).
		SetRetryCount(configs.RetryDuration).
		SetRetryWaitTime(time.Duration(configs.RetryInterval) * time.Second).
		AddRetryCondition(
			func(r *resty.Response, err error) bool {
				if err != nil {
					return true
				}
				return r.StatusCode() >= 500 || r.StatusCode() == http.StatusTooManyRequests
			},
		)
	event := audit.NewEvent(logger.Log.Sugar())
	var validAudit bool
	if configs.ValidFileAudit {
		sub := audit.NewFileSubscriber(configs.AuditFile, logger.Log.Sugar())
		if sub != nil {
			event.Register(sub)
			defer func() {
				if err := sub.Close(); err != nil {
					logger.Log.Sugar().Errorw("ошибка при закрытии файла аудита", "err", err)
				}
			}()
			validAudit = true
		}
	}
	if configs.ValidURLAudit {
		sub := audit.NewURLSubscriber(configs.AuditURL, logger.Log.Sugar(), client)
		if sub != nil {
			event.Register(sub)
			defer sub.Close()
			validAudit = true
		}
	}
	h := handler.NewMetricsHandler(storage, logger.Log.Sugar())
	r := chi.NewRouter()
	r.Use(handler.RequestLogger(logger.Log))
	r.Use(handler.GzipMiddleware)
	r.Use(handler.HashMiddleware(configs.Key))
	r.Get("/value/{mType}/{mName}", h.GetMetricHandler())
	r.Post("/value", h.APIValueHandler())
	r.Post("/value/", h.APIValueHandler())
	r.Get("/ping", h.PingHandler())
	r.Get("/", h.ExposeMetricsHandler())
	r.Group(func(r chi.Router) {
		if validAudit {
			r.Use(handler.AuditMiddleware(event))
		}
		r.Post("/update/{mType}/{mName}/{mVal}", h.PostMetricsHandler())
		r.Post("/updates", h.APIUpdatesHandler())
		r.Post("/updates/", h.APIUpdatesHandler())
		r.Post("/update", h.APIUpdateHandler())
		r.Post("/update/", h.APIUpdateHandler())
	})
	go func() {
		if err := http.ListenAndServe(":8081", nil); err != nil {
			logger.Log.Sugar().Errorw("ошибка запуска pprof сервера", "error", err)
		}
	}()
	serv := server.New(configs.Net.String(), logger.Log.Sugar(), r)
	if err := serv.Run(); err != nil {
		panic(err)
	}
}
