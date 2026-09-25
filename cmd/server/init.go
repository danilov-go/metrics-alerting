package main

import (
	"crypto/rsa"
	"net"
	"net/http"
	"time"

	pb "github.com/danilov-go/metrics-alerting.git/internal/proto"
	"github.com/danilov-go/metrics-alerting.git/internal/server"
	"go.uber.org/zap"

	"github.com/danilov-go/metrics-alerting.git/internal/audit"
	"github.com/danilov-go/metrics-alerting.git/internal/config"
	"github.com/danilov-go/metrics-alerting.git/internal/config/db"
	"github.com/danilov-go/metrics-alerting.git/internal/handler"
	"github.com/danilov-go/metrics-alerting.git/internal/logger"
	"github.com/danilov-go/metrics-alerting.git/internal/repository"
	"github.com/go-chi/chi/v5"
	"github.com/go-resty/resty/v2"
	"google.golang.org/grpc"
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

func initRouter(configs config.ConfigServer, storage handler.Storage, event *audit.Event, fileSub *audit.FileSubscriber, urlSub *audit.URLSubscriber, ipNet *net.IPNet, privatKey *rsa.PrivateKey) chi.Router {
	h := handler.NewMetricsHandler(storage, logger.Log.Sugar())
	r := chi.NewRouter()
	public := func(router chi.Router) {
		router.Use(handler.RequestLogger(logger.Log))
		if configs.CryptoKey != "" {
			router.Use(handler.CryptoMiddleware(privatKey))
		}
		router.Use(handler.GzipMiddleware)
		router.Use(handler.HashMiddleware(configs.Key))
	}
	r.Group(func(pub chi.Router) {
		public(pub)
		pub.Get("/value/{mType}/{mName}", h.GetMetricHandler())
		pub.Post("/value", h.APIValueHandler())
		pub.Post("/value/", h.APIValueHandler())
		pub.Get("/ping", h.PingHandler())
		pub.Get("/", h.ExposeMetricsHandler())
	})
	r.Group(func(pub chi.Router) {
		if ipNet != nil {
			pub.Use(handler.TrustedMiddleware(ipNet))
		}
		if fileSub != nil || urlSub != nil {
			pub.Use(handler.AuditMiddleware(event))
		}
		public(pub)
		pub.Post("/update/{mType}/{mName}/{mVal}", h.PostMetricsHandler())
		pub.Post("/updates", h.APIUpdatesHandler())
		pub.Post("/updates/", h.APIUpdatesHandler())
		pub.Post("/update", h.APIUpdateHandler())
		pub.Post("/update/", h.APIUpdateHandler())
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

func runServerPprof() {
	go func() {
		if err := http.ListenAndServe(":8081", nil); err != nil {
			logger.Log.Sugar().Errorw("ошибка запуска pprof сервера", "err", err)
		}
	}()
}

func initGRPC(configs config.ConfigServer, iPNet *net.IPNet, storage handler.Storage, l *zap.Logger, privatKey *rsa.PrivateKey) (*server.GRPCServer, error) {
	if configs.GrpcAddress == "" {
		return nil, nil
	}
	listen, err := net.Listen("tcp", configs.GrpcAddress)
	if err != nil {
		return nil, err
	}
	gServ := grpc.NewServer(
		grpc.UnaryInterceptor(handler.UnaryInterceptor(iPNet)),
	)
	h := handler.NewGRPCHandler(storage, l.Sugar(), privatKey)
	pb.RegisterMetricsServer(gServ, h)
	serv := server.NewGRPC(configs.GrpcAddress, l.Sugar(), gServ, listen)
	return serv, nil
}
