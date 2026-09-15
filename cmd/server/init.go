package main

import (
	"net"
	"net/http"
	"time"

	"github.com/danilov-go/metrics-alerting.git/internal/audit"
	"github.com/danilov-go/metrics-alerting.git/internal/config"
	"github.com/danilov-go/metrics-alerting.git/internal/config/db"
	"github.com/danilov-go/metrics-alerting.git/internal/handler"
	"github.com/danilov-go/metrics-alerting.git/internal/logger"
	"github.com/danilov-go/metrics-alerting.git/internal/repository"
	"github.com/go-chi/chi/v5"
	"github.com/go-resty/resty/v2"
)

func initStorage(configs config.ConfigServer) handler.Storage {
	var pg handler.Storage
	var dbErr error
	if configs.DatabaseDSN != "" {
		pg, dbErr = db.InitDB(configs.DatabaseDSN)
	}
	if dbErr == nil && pg != nil {
		duration := time.Duration(configs.RetryDuration) * time.Second
		interval := time.Duration(configs.RetryInterval) * time.Second
		return handler.NewErrorMiddleware(pg, duration, interval)
	}
	cfg := repository.ConfigFile{
		Path:     configs.FileStoragePath,
		Interval: time.Duration(configs.StoreIntrval) * time.Second,
		Restore:  configs.Restore,
	}
	return repository.InitMemStorage(cfg, logger.Log.Sugar())
}

func initRouter(configs config.ConfigServer, storage handler.Storage, event *audit.Event, fileSub *audit.FileSubscriber, urlSub *audit.URLSubscriber) chi.Router {
	h := handler.NewMetricsHandler(storage, logger.Log.Sugar())
	r := chi.NewRouter()
	if configs.TrustedSubnet != "" {
		_, ipNet, err := net.ParseCIDR(configs.TrustedSubnet)
		if err != nil {
			logger.Log.Sugar().Fatal("ошибка парсинга TrustedSubnet", "error", err)
		}
		r.Use(handler.TrustedMiddleware(ipNet))
	}
	r.Use(handler.RequestLogger(logger.Log))
	if configs.CryptoKey != "" {
		privatKey, err := configs.GetKey()
		if err != nil {
			logger.Log.Sugar().Fatal("ошибка загрузки или парсинга приватного ключа сервера")
		}
		r.Use(handler.CryptoMiddleware(privatKey))
	}
	r.Use(handler.GzipMiddleware)
	r.Use(handler.HashMiddleware(configs.Key))
	r.Get("/value/{mType}/{mName}", h.GetMetricHandler())
	r.Post("/value", h.APIValueHandler())
	r.Post("/value/", h.APIValueHandler())
	r.Get("/ping", h.PingHandler())
	r.Get("/", h.ExposeMetricsHandler())
	r.Group(func(r chi.Router) {
		if fileSub != nil || urlSub != nil {
			r.Use(handler.AuditMiddleware(event))
		}
		r.Post("/update/{mType}/{mName}/{mVal}", h.PostMetricsHandler())
		r.Post("/updates", h.APIUpdatesHandler())
		r.Post("/updates/", h.APIUpdatesHandler())
		r.Post("/update", h.APIUpdateHandler())
		r.Post("/update/", h.APIUpdateHandler())
	})
	return r
}

func initClient(configs config.ConfigServer) *resty.Client {
	client := resty.New().
		SetTimeout(5 * time.Second).
		SetRetryCount(int(configs.RetryDuration)).
		SetRetryWaitTime(time.Duration(configs.RetryInterval) * time.Second).
		AddRetryCondition(
			func(r *resty.Response, err error) bool {
				if err != nil {
					return true
				}
				return r.StatusCode() >= 500 || r.StatusCode() == http.StatusTooManyRequests
			},
		)
	return client
}

func initAudit(configs config.ConfigServer, client *resty.Client) (*audit.Event, *audit.FileSubscriber, *audit.URLSubscriber) {
	event := audit.NewEvent(logger.Log.Sugar())
	var fileSub *audit.FileSubscriber
	var urlSub *audit.URLSubscriber
	if configs.AuditFile != "" {
		fileSub = audit.NewFileSubscriber(configs.AuditFile, logger.Log.Sugar())
		if fileSub != nil {
			event.Register(fileSub)
		}
	}
	if configs.AuditURL != "" {
		urlSub = audit.NewURLSubscriber(configs.AuditURL, logger.Log.Sugar(), client)
		if urlSub != nil {
			event.Register(urlSub)
		}
	}
	return event, fileSub, urlSub
}

func runProffServer() {
	go func() {
		if err := http.ListenAndServe(":8081", nil); err != nil {
			logger.Log.Sugar().Errorw("ошибка запуска pprof сервера", "err", err)
		}
	}()
}
