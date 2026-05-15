package main

import (
	"github.com/danilov-go/metrics-alerting.git/internal/repository"
	"github.com/danilov-go/metrics-alerting.git/internal/server"
)

func main() {
	metricsStorage := &repository.MemStorage{
		Gauges:   make(map[string]float64),
		Counters: make(map[string]int64),
	}
	serv := server.New(metricsStorage)
	if err := serv.Run(); err != nil {
		panic(err)
	}
}
