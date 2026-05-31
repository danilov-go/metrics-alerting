package repository

import (
	"bufio"
	"encoding/json"
	"os"
	"sync"
	"time"

	"github.com/danilov-go/metrics-alerting.git/internal/logger"
	"go.uber.org/zap"
)

type MemStorage struct {
	Gauges   map[string]float64 `json:"gauges"`
	Counters map[string]int64   `json:"counters"`
	mu       sync.RWMutex       `json:"-"`
}

func (g *MemStorage) SaveGauges(name string, value float64) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.Gauges == nil {
		g.Gauges = make(map[string]float64)
	}
	g.Gauges[name] = value
}

func (c *MemStorage) SaveCounters(name string, value int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.Counters == nil {
		c.Counters = make(map[string]int64)
	}
	if val, ok := c.Counters[name]; ok {
		c.Counters[name] = value + val
	} else {
		c.Counters[name] = value
	}
}

func (g *MemStorage) GetGauges(name string) (float64, bool) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.Gauges == nil {
		g.Gauges = make(map[string]float64)
	}
	val, ok := g.Gauges[name]
	return val, ok
}

func (c *MemStorage) GetCounters(name string) (int64, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.Counters == nil {
		c.Counters = make(map[string]int64)
	}
	val, ok := c.Counters[name]
	return val, ok
}

func (g *MemStorage) GetAllGauges() map[string]float64 {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.Gauges == nil {
		g.Gauges = make(map[string]float64)
	}
	return g.Gauges
}

func (c *MemStorage) GetAllCounters() map[string]int64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.Counters == nil {
		c.Counters = make(map[string]int64)
	}
	return c.Counters
}

func (m *MemStorage) SaveFile(path string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer file.Close()
	buf := bufio.NewWriter(file)
	err = json.NewEncoder(buf).Encode(m)
	if err != nil {
		return err

	}
	return buf.Flush()
}

func (m *MemStorage) LoadFile(path string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	file, err := os.OpenFile(path, os.O_RDONLY, 0666)
	if err != nil {
		return err
	}
	defer file.Close()
	buf := bufio.NewReader(file)
	err = json.NewDecoder(buf).Decode(m)
	if err != nil {
		return err
	}
	return nil
}

func (m *MemStorage) Run(path string, storeIntrval time.Duration) error {
	ticker := time.NewTicker(storeIntrval)
	go func(t *time.Ticker) {
		defer t.Stop()
		for range t.C {
			err := m.SaveFile(path)
			if err != nil {
				logger.Log.Info("ошибка записи в файл", zap.Error(err))
			}
		}
	}(ticker)
	return nil
}
