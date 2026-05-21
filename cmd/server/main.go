package main

import (
	"github.com/danilov-go/metrics-alerting.git/internal/config"
	"github.com/danilov-go/metrics-alerting.git/internal/handler"
	"github.com/danilov-go/metrics-alerting.git/internal/repository"
	"github.com/danilov-go/metrics-alerting.git/internal/server"
	"github.com/go-chi/chi/v5"
)

func main() {
	configs := config.ConfigServer{
		Net: config.NetAddress{
			Host: "localhost",
			Port: 8080,
		},
	}
	configs.Get()
	metricsStorage := &repository.MemStorage{
		Gauges:   make(map[string]float64),
		Counters: make(map[string]int64),
	}
	r := chi.NewRouter()
	r.Post("/update/{mType}/{mName}/{mVal}", handler.PostMetricsHandler(metricsStorage))
	r.Get("/value/{mType}/{mName}", handler.GetMetricHandler(metricsStorage))
	r.Get("/", handler.ExposeMetricsHandler(metricsStorage))
	serv := server.New(configs.Net.String(), r)
	if err := serv.Run(); err != nil {
		panic(err)
	}
}
