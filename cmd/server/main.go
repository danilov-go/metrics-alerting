package main

import (
	"github.com/danilov-go/metrics-alerting.git/internal/config"
	"github.com/danilov-go/metrics-alerting.git/internal/repository"
	"github.com/danilov-go/metrics-alerting.git/internal/server"
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
	serv := server.New(configs.Net.String(), metricsStorage)
	if err := serv.Run(); err != nil {
		panic(err)
	}
}
