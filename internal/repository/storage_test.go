package repository_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/danilov-go/metrics-alerting.git/internal/models"
	"github.com/danilov-go/metrics-alerting.git/internal/repository"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestMemStorage_SaveGauges(t *testing.T) {
	ctx := context.Background()
	logger := zaptest.NewLogger(t)
	storage := repository.InitMemStorage(repository.ConfigFile{}, logger.Sugar())
	tests := []struct {
		name  string
		mName string
		mVal  float64
	}{
		{
			name:  "положительный тест",
			mName: "Alloc",
			mVal:  123.45,
		},
		{
			name:  "обновление метрики",
			mName: "Alloc",
			mVal:  223.45,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := storage.SaveGauges(ctx, tt.mName, tt.mVal)
			assert.NoError(t, err)
			val, ok := storage.Gauges[tt.mName]
			assert.True(t, ok)
			assert.Equal(t, tt.mVal, val)
		})
	}
}

func TestMemStorage_SaveCounters(t *testing.T) {
	ctx := context.Background()
	logger := zaptest.NewLogger(t)
	storage := repository.InitMemStorage(repository.ConfigFile{}, logger.Sugar())
	tests := []struct {
		name   string
		mName  string
		mVal   int64
		expVal int64
	}{
		{
			name:   "положительный тест",
			mName:  "PollCount",
			mVal:   5,
			expVal: 5,
		},
		{
			name:   "обновление метрики",
			mName:  "PollCount",
			mVal:   5,
			expVal: 10,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := storage.SaveCounters(ctx, tt.mName, tt.mVal)
			assert.NoError(t, err)
			val, ok := storage.Counters[tt.mName]
			assert.True(t, ok)
			assert.Equal(t, tt.expVal, val)

		})
	}
}

func TestMemStorage_GetGauges(t *testing.T) {
	ctx := context.Background()
	logger := zaptest.NewLogger(t)
	storage := repository.InitMemStorage(repository.ConfigFile{}, logger.Sugar())
	tests := []struct {
		name    string
		mName   string
		mVal    float64
		wantErr error
	}{
		{
			name:    "отсутствие метрик",
			mName:   "Alloc",
			mVal:    0,
			wantErr: sql.ErrNoRows,
		},
		{
			name:    "положительный тест",
			mName:   "Alloc",
			mVal:    123.45,
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr == nil {
				err := storage.SaveGauges(ctx, tt.mName, tt.mVal)
				assert.NoError(t, err)
			}
			val, err := storage.GetGauges(ctx, tt.mName)
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.mVal, val)
		})
	}
}

func TestMemStorage_GetCounters(t *testing.T) {
	ctx := context.Background()
	logger := zaptest.NewLogger(t)
	storage := repository.InitMemStorage(repository.ConfigFile{}, logger.Sugar())
	tests := []struct {
		name    string
		mName   string
		mVal    int64
		wantErr error
	}{
		{
			name:    "отсутствие метрик",
			mName:   "PollCount",
			mVal:    0,
			wantErr: sql.ErrNoRows,
		},
		{
			name:    "положительный тест",
			mName:   "PollCount",
			mVal:    5,
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr == nil {
				err := storage.SaveCounters(ctx, tt.mName, tt.mVal)
				assert.NoError(t, err)
			}
			val, err := storage.GetCounters(ctx, tt.mName)
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.mVal, val)
		})
	}
}

func TestMemStorage_SaveAll(t *testing.T) {
	ctx := context.Background()
	logger := zaptest.NewLogger(t)
	tests := []struct {
		name         string
		metricsExp   []models.Metrics
		wantErr      bool
		errMsg       string
		wantGauges   map[string]float64
		wantCounters map[string]int64
	}{
		{
			name: "положительный тест",
			metricsExp: []models.Metrics{
				{ID: "Alloc", MType: models.Gauge, Value: models.PointerFloat64(123.45)},
				{ID: "CPUutilization1", MType: models.Gauge, Value: models.PointerFloat64(38.88)},
				{ID: "CPUutilization2", MType: models.Gauge, Value: models.PointerFloat64(35.17)},
				{ID: "CPUutilization3", MType: models.Gauge, Value: models.PointerFloat64(22.22)},
				{ID: "CPUutilization4", MType: models.Gauge, Value: models.PointerFloat64(18.50)},
				{ID: "CPUutilization5", MType: models.Gauge, Value: models.PointerFloat64(0.99)},
				{ID: "CPUutilization6", MType: models.Gauge, Value: models.PointerFloat64(0.50)},
				{ID: "CPUutilization7", MType: models.Gauge, Value: models.PointerFloat64(0.49)},
				{ID: "CPUutilization8", MType: models.Gauge, Value: models.PointerFloat64(0.99)},
				{ID: "PollCount", MType: models.Counter, Delta: models.PointerInt64(5)},
			},
			wantErr: false,
			wantGauges: map[string]float64{
				"Alloc":           123.45,
				"CPUutilization1": 38.88,
				"CPUutilization2": 35.17,
				"CPUutilization3": 22.22,
				"CPUutilization4": 18.50,
				"CPUutilization5": 0.99,
				"CPUutilization6": 0.50,
				"CPUutilization7": 0.49,
				"CPUutilization8": 0.99,
			},
			wantCounters: map[string]int64{
				"PollCount": 5,
			},
		},
		{
			name: "пустой delta",
			metricsExp: []models.Metrics{
				{ID: "PollCount", MType: models.Counter, Delta: nil},
			},
			wantErr: true,
			errMsg:  "переменная delta пустая",
		},
		{
			name: "пустой value",
			metricsExp: []models.Metrics{
				{ID: "Alloc", MType: models.Gauge, Value: nil},
			},
			wantErr: true,
			errMsg:  "переменная value пустая",
		},
		{
			name: "неизвестный тип",
			metricsExp: []models.Metrics{
				{ID: "unknown", MType: "unknown"},
			},
			wantErr: true,
			errMsg:  "неизвестный тип метрики",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := repository.InitMemStorage(repository.ConfigFile{}, logger.Sugar())
			err := storage.SaveAll(ctx, tt.metricsExp)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantGauges, storage.Gauges)
				assert.Equal(t, tt.wantCounters, storage.Counters)
			}
		})
	}
}

func TestMemStorage_GetAllGauges(t *testing.T) {
	ctx := context.Background()
	logger := zaptest.NewLogger(t)
	tests := []struct {
		name       string
		metricsExp []models.Metrics
		wantGauges map[string]float64
	}{
		{
			name:       "пустое хранилище",
			metricsExp: []models.Metrics{},
			wantGauges: map[string]float64{},
		},
		{
			name: "положительный тест",
			metricsExp: []models.Metrics{
				{ID: "Alloc", MType: models.Gauge, Value: models.PointerFloat64(123.45)},
				{ID: "CPUutilization1", MType: models.Gauge, Value: models.PointerFloat64(38.88)},
				{ID: "CPUutilization2", MType: models.Gauge, Value: models.PointerFloat64(35.17)},
				{ID: "CPUutilization3", MType: models.Gauge, Value: models.PointerFloat64(22.22)},
				{ID: "CPUutilization4", MType: models.Gauge, Value: models.PointerFloat64(18.50)},
				{ID: "CPUutilization5", MType: models.Gauge, Value: models.PointerFloat64(0.99)},
				{ID: "CPUutilization6", MType: models.Gauge, Value: models.PointerFloat64(0.50)},
				{ID: "CPUutilization7", MType: models.Gauge, Value: models.PointerFloat64(0.49)},
				{ID: "CPUutilization8", MType: models.Gauge, Value: models.PointerFloat64(0.99)},
			},
			wantGauges: map[string]float64{
				"Alloc":           123.45,
				"CPUutilization1": 38.88,
				"CPUutilization2": 35.17,
				"CPUutilization3": 22.22,
				"CPUutilization4": 18.50,
				"CPUutilization5": 0.99,
				"CPUutilization6": 0.50,
				"CPUutilization7": 0.49,
				"CPUutilization8": 0.99,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := repository.InitMemStorage(repository.ConfigFile{}, logger.Sugar())

			if len(tt.metricsExp) > 0 {
				err := storage.SaveAll(ctx, tt.metricsExp)
				assert.NoError(t, err)
			}
			metrics, err := storage.GetAllGauges(ctx)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantGauges, metrics)
		})
	}
}

func TestMemStorage_GetAllCounters(t *testing.T) {
	ctx := context.Background()
	logger := zaptest.NewLogger(t)
	tests := []struct {
		name         string
		metricsExp   []models.Metrics
		wantCounters map[string]int64
	}{
		{
			name:         "пустое хранилище",
			metricsExp:   []models.Metrics{},
			wantCounters: map[string]int64{},
		},
		{
			name: "положительный тест",
			metricsExp: []models.Metrics{
				{ID: "PollCount", MType: models.Counter, Delta: models.PointerInt64(5)},
				{ID: "PollCountTest", MType: models.Counter, Delta: models.PointerInt64(10)},
			},
			wantCounters: map[string]int64{
				"PollCount":     5,
				"PollCountTest": 10,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := repository.InitMemStorage(repository.ConfigFile{}, logger.Sugar())
			if len(tt.metricsExp) > 0 {
				err := storage.SaveAll(ctx, tt.metricsExp)
				assert.NoError(t, err)
			}
			metrics, err := storage.GetAllCounters(ctx)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantCounters, metrics)
		})
	}
}

func TestMemStorage_Ping(t *testing.T) {
	logger := zaptest.NewLogger(t)
	tests := []struct {
		name      string
		cancelCtx bool
		wantErr   error
	}{
		{
			name:      "положительный тест",
			cancelCtx: false,
			wantErr:   errors.New("ошибка подключения к БД"),
		},
		{
			name:      "ошибка контекста",
			cancelCtx: true,
			wantErr:   context.Canceled,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := repository.InitMemStorage(repository.ConfigFile{}, logger.Sugar())
			var ctx context.Context
			var cancel context.CancelFunc
			if tt.cancelCtx {
				ctx, cancel = context.WithCancel(context.Background())
				cancel()
			} else {
				ctx = context.Background()
			}
			err := storage.Ping(ctx)
			assert.Error(t, err)
			if tt.cancelCtx {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.Equal(t, tt.wantErr.Error(), err.Error())
			}
		})
	}
}
