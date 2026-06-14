package repository

import (
	"bufio"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
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
	logger   log
	filePath string        `json:"-"`
	interval time.Duration `json:"-"`
}

func InitMemStorage(cfg ConfigFile, l log) *MemStorage {
	m := &MemStorage{
		Gauges:   make(map[string]float64),
		Counters: make(map[string]int64),
		logger:   l,
		filePath: cfg.Path,
		interval: cfg.Interval,
	}
	if cfg.Path != "" && cfg.Restore {
		err := m.loadFile(cfg.Path)
		if err != nil {
			l.Infow("ошибка загрузки метрик из файла", "error", err)
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
	defer c.mu.Unlock()
	if c.Counters == nil {
		c.Counters = make(map[string]int64)
	}
	if val, ok := c.Counters[name]; ok {
		c.Counters[name] = value + val
	} else {
		c.Counters[name] = value
	}
	if c.interval == 0 {
		if err := c.save(); err != nil {
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
		return 0, sql.ErrNoRows
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
		return 0, sql.ErrNoRows
	}
	return val, nil
}

func (g *MemStorage) GetAllGauges(ctx context.Context) (map[string]float64, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.Gauges, nil
}

func (c *MemStorage) GetAllCounters(ctx context.Context) (map[string]int64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.Counters, nil
}

func (m *MemStorage) SaveAll(ctx context.Context, metrics []models.Metrics) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, metric := range metrics {
		switch metric.MType {
		case models.Counter:
			if metric.Delta == nil {
				return errors.New("переменная delta пустая")
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
				return errors.New("переменная value пустая")
			}
			if m.Gauges == nil {
				m.Gauges = make(map[string]float64)
			}
			m.Gauges[metric.ID] = *metric.Value
		default:
			return errors.New("неизвестный тип метрики")
		}
	}
	if m.interval == 0 {
		if err := m.save(); err != nil {
			return err
		}
	}
	return nil
}

func (m *MemStorage) Ping(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return errors.New("ошибка подключения к БД")
}

func (m *MemStorage) SaveFile() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.save()
}

func (m *MemStorage) save() error {
	if m.filePath == "" {
		return nil
	}
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

func (m *MemStorage) run() {
	ticker := time.NewTicker(m.interval)
	go func(t *time.Ticker) {
		defer t.Stop()
		for range t.C {
			err := m.SaveFile()
			if err != nil {
				m.logger.Errorw("ошибка сохранения в файл", "error", err)
			}
		}
	}(ticker)
}
