package main

import (
	"time"

	"github.com/danilov-go/metrics-alerting.git/internal/agent"
)

const (
	pollInterval   = 2
	reportInterval = 10
)

func main() {
	step := reportInterval / pollInterval
	client := agent.New("")
	var pollCount int64 = 0
	for {
		client.Run(&pollCount, step)
		time.Sleep(pollInterval * time.Second)
	}
}
