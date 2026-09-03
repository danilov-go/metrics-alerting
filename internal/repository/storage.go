// Package repository реализует хранилище данных в оперативной памяти.
package repository

import (
	"bufio"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"maps"
	"os"
	"sync"
	"time"

	"github.com/danilov-go/metrics-alerting.git/internal/models"
)

type log interface {
	Errorw(msg string, keysAndValues ...any)
	Infow(msg string, keysAndValues ...any)
}

// ConfigFile опредиляет параметры для конфигурации файлового хранилища.
type ConfigFile struct {
	// Path определяет путь к файлу для сохранения данных.
	Path string
	// Interval определяет периодичность записи метрик в файл.
	Interval time.Duration
	// Restore определяет необходимость загрузки ранее сохраненных метрик при старте.
	Restore bool
}

// MemStorage реализует хранилище данных в оперативной памяти.
type MemStorage struct {
	// Gauges хранит метрики типа gauge.
	Gauges map[string]float64 `json:"gauges"`
	// Counters хранит метрики типа counter.
	Counters map[string]int64 `json:"counters"`
	mu       sync.RWMutex     `json:"-"`
	logger   log
	filePath string        `json:"-"`
	interval time.Duration `json:"-"`
}

// InitMemStorage создает новый экземпляр MemStorage.
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

// SaveGauges сохраняет или перезаписывает метрику типа "gauge".
func (m *MemStorage) SaveGauges(ctx context.Context, name string, value float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.Gauges == nil {
		m.Gauges = make(map[string]float64)
	}
	m.Gauges[name] = value
	return nil
}

// SaveCounters сохраняет или обновляет метрику типа "counter".
func (m *MemStorage) SaveCounters(ctx context.Context, name string, value int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.Counters == nil {
		m.Counters = make(map[string]int64)
	}
	if val, ok := m.Counters[name]; ok {
		m.Counters[name] = value + val
	} else {
		m.Counters[name] = value
	}
	if m.interval == 0 {
		if err := m.save(); err != nil {
			return err
		}
	}
	return nil
}

// GetGauges возвращает значение метрики типа "gauge" по её названию.
func (m *MemStorage) GetGauges(ctx context.Context, name string) (float64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.Gauges == nil {
		m.Gauges = make(map[string]float64)
	}
	val, ok := m.Gauges[name]
	if !ok {
		return 0, sql.ErrNoRows
	}
	return val, nil
}

// GetCounters возвращает значение метрики типа "counter" по её названию.
func (m *MemStorage) GetCounters(ctx context.Context, name string) (int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.Counters == nil {
		m.Counters = make(map[string]int64)
	}
	val, ok := m.Counters[name]
	if !ok {
		return 0, sql.ErrNoRows
	}
	return val, nil
}

// GetAllGauges возвращает map всех хранящихся метрик типа "gauge".
func (m *MemStorage) GetAllGauges(ctx context.Context) (map[string]float64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	gauges := make(map[string]float64, len(m.Gauges))
	maps.Copy(gauges, m.Gauges)
	return gauges, nil
}

// GetAllCounters возвращает map всех хранящихся метрик типа "counter".
func (m *MemStorage) GetAllCounters(ctx context.Context) (map[string]int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	counters := make(map[string]int64, len(m.Counters))
	maps.Copy(counters, m.Counters)
	return counters, nil
}

// SaveAll выполняет пакетное сохранение метрик.
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

// Ping проверяет доступность хранилища.
func (m *MemStorage) Ping(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return errors.New("ошибка подключения к БД")
}

// SaveFile сохраняет текущие метрики в файл.
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
	defer func() { _ = file.Close() }()
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
	defer func() { _ = file.Close() }()
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
