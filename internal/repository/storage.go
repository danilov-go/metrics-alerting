package repository

import (
	"bufio"
	"encoding/json"
	"os"
	"sync"
	"time"
)

type log interface {
	Errorw(msg string, keysAndValues ...any)
	Infow(msg string, keysAndValues ...any)
}
type ConfigFile struct {
	Path     string
	Interval time.Duration
	Restore  bool
}

type MemStorage struct {
	Gauges   map[string]float64 `json:"gauges"`
	Counters map[string]int64   `json:"counters"`
	mu       sync.RWMutex       `json:"-"`
	Logger   log                `json:"-"`
	filePath string             `json:"-"`
	interval time.Duration      `json:"-"`
}

func InitMemStorage(l log, cfg ConfigFile) *MemStorage {
	m := &MemStorage{
		Gauges:   make(map[string]float64),
		Counters: make(map[string]int64),
		Logger:   l,
		filePath: cfg.Path,
		interval: cfg.Interval,
	}
	if cfg.Path != "" && cfg.Restore {
		err := m.loadFile(cfg.Path)
		if err != nil {
			l.Infow("не удалось восстановить метрики из файла", "err", err)
		}
	}
	if cfg.Path != "" && cfg.Interval > 0 {
		m.run()
	}
	return m
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
	if c.Counters == nil {
		c.Counters = make(map[string]int64)
	}
	if val, ok := c.Counters[name]; ok {
		c.Counters[name] = value + val
	} else {
		c.Counters[name] = value
	}
	c.mu.Unlock()
	if c.interval == 0 {
		if err := c.SaveFile(); err != nil {
			c.Logger.Errorw("ошибка синхронной записи", "err", err)
		}
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

func (m *MemStorage) SaveFile() error {
	if m.filePath == "" {
		return nil
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	file, err := os.OpenFile(m.filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
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

func (m *MemStorage) loadFile(path string) error {
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

func (m *MemStorage) run() error {
	ticker := time.NewTicker(m.interval)
	go func(t *time.Ticker) {
		defer t.Stop()
		for range t.C {
			err := m.SaveFile()
			if err != nil {
				m.Logger.Infow("ошибка записи в файл", "err", err)
			}
		}
	}(ticker)
	return nil
}
