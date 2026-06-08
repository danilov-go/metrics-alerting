package repository

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/danilov-go/metrics-alerting.git/internal/models"
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

func (g *MemStorage) SaveGauges(ctx context.Context, name string, value float64) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.Gauges == nil {
		g.Gauges = make(map[string]float64)
	}
	g.Gauges[name] = value
	return nil
}

func (c *MemStorage) SaveCounters(ctx context.Context, name string, value int64) error {
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
			return err
		}
	}
	return nil
}

func (g *MemStorage) GetGauges(ctx context.Context, name string) (float64, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.Gauges == nil {
		g.Gauges = make(map[string]float64)
	}
	val, ok := g.Gauges[name]
	if !ok {
		return 0, fmt.Errorf("метрики gauge отсутствуют")
	}
	return val, nil
}

func (c *MemStorage) GetCounters(ctx context.Context, name string) (int64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.Counters == nil {
		c.Counters = make(map[string]int64)
	}
	val, ok := c.Counters[name]
	if !ok {
		return 0, fmt.Errorf("метрики counter отсутствуют")
	}
	return val, nil
}

func (g *MemStorage) GetAllGauges(ctx context.Context) (map[string]float64, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.Gauges == nil {
		return nil, fmt.Errorf("метрики gauge отсутствуют")
	}
	return g.Gauges, nil
}

func (c *MemStorage) GetAllCounters(ctx context.Context) (map[string]int64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.Counters == nil {
		return nil, fmt.Errorf("метрики counter отсутствуют")
	}
	return c.Counters, nil
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

func (m *MemStorage) SaveAll(ctx context.Context, metrics []models.Metrics) error {
	m.mu.Lock()
	for _, metric := range metrics {
		switch metric.MType {
		case models.Counter:
			if metric.Delta == nil {
				m.mu.Unlock()
				return fmt.Errorf("переменная delta пустая")
			}
			if m.Counters == nil {
				m.Counters = make(map[string]int64)
			}
			if val, ok := m.Counters[metric.ID]; ok {
				m.Counters[metric.ID] = *metric.Delta + val
			} else {
				m.Counters[metric.ID] = *metric.Delta
			}
		case models.Gauge:
			if metric.Value == nil {
				m.mu.Unlock()
				return fmt.Errorf("переменная value пустая")
			}
			if m.Gauges == nil {
				m.Gauges = make(map[string]float64)
			}
			m.Gauges[metric.ID] = *metric.Value
		default:
			m.mu.Unlock()
			return fmt.Errorf("неизвестный тип метрики")
		}
	}
	m.mu.Unlock()
	if m.interval == 0 {
		if err := m.SaveFile(); err != nil {
			m.Logger.Errorw("ошибка синхронной записи", "err", err)
			return err
		}
	}
	return nil
}
