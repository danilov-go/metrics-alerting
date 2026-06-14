package handler

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/danilov-go/metrics-alerting.git/internal/models"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

type log interface {
	Errorw(msg string, keysAndValues ...any)
}

type MetricsHandler struct {
	storage Storage
	logger  log
}

func NewMetricsHandler(storage Storage, l log) *MetricsHandler {
	return &MetricsHandler{
		storage: storage,
		logger:  l,
	}
}

type Storage interface {
	SaveCounters(ctx context.Context, name string, delta int64) error
	SaveGauges(ctx context.Context, name string, value float64) error
	GetGauges(ctx context.Context, name string) (float64, error)
	GetCounters(ctx context.Context, name string) (int64, error)
	GetAllGauges(ctx context.Context) (map[string]float64, error)
	GetAllCounters(ctx context.Context) (map[string]int64, error)
	SaveAll(ctx context.Context, metrics []models.Metrics) error
	Ping(ctx context.Context) error
}

type PGErrorClassification int

const (
	NonRetriable PGErrorClassification = iota
	Retriable
)

type PostgresErrorClassifier struct{}

func NewPostgresErrorClassifier() *PostgresErrorClassifier {
	return &PostgresErrorClassifier{}
}

func (c *PostgresErrorClassifier) Classify(err error) PGErrorClassification {
	if err == nil {
		return NonRetriable
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return classifyPgError(pgErr)
	}
	return NonRetriable
}

func classifyPgError(pgErr *pgconn.PgError) PGErrorClassification {
	switch pgErr.Code {
	case pgerrcode.ConnectionException,
		pgerrcode.ConnectionDoesNotExist,
		pgerrcode.ConnectionFailure:
		return Retriable
	}
	return NonRetriable
}

type ErrorStorageMiddleware struct {
	next       Storage
	classifier *PostgresErrorClassifier
}

func NewErrorMiddleware(next Storage) *ErrorStorageMiddleware {
	return &ErrorStorageMiddleware{
		next:       next,
		classifier: NewPostgresErrorClassifier(),
	}
}

func (rm *ErrorStorageMiddleware) replay(ctx context.Context, operation func() error) error {
	const maxRetries = 3
	duration := 1
	for attempt := 0; attempt <= maxRetries; attempt++ {
		err := operation()
		if err == nil {
			return nil
		}
		if attempt == maxRetries {
			return fmt.Errorf("попытки подключения исчерпаны: %w", err)
		}
		if rm.classifier.Classify(err) == Retriable {
			timer := time.NewTimer(time.Duration(duration) * time.Second)
			select {
			case <-ctx.Done():
				timer.Stop()
				return ctx.Err()
			case <-timer.C:
			}
			timer.Stop()
			duration += 2
			continue
		}
		return err
	}
	return nil
}

func (rm *ErrorStorageMiddleware) SaveCounters(ctx context.Context, name string, delta int64) error {
	operation := func() error {
		return rm.next.SaveCounters(ctx, name, delta)
	}
	return rm.replay(ctx, operation)
}

func (rm *ErrorStorageMiddleware) SaveGauges(ctx context.Context, name string, value float64) error {
	operation := func() error {
		return rm.next.SaveGauges(ctx, name, value)
	}
	return rm.replay(ctx, operation)
}

func (rm *ErrorStorageMiddleware) GetGauges(ctx context.Context, name string) (float64, error) {
	var value float64
	operation := func() error {
		var err error
		value, err = rm.next.GetGauges(ctx, name)
		return err
	}
	err := rm.replay(ctx, operation)
	return value, err
}

func (rm *ErrorStorageMiddleware) GetCounters(ctx context.Context, name string) (int64, error) {
	var delta int64
	operation := func() error {
		var err error
		delta, err = rm.next.GetCounters(ctx, name)
		return err
	}
	err := rm.replay(ctx, operation)
	return delta, err
}

func (rm *ErrorStorageMiddleware) GetAllGauges(ctx context.Context) (map[string]float64, error) {
	var gauges map[string]float64
	operation := func() error {
		var err error
		gauges, err = rm.next.GetAllGauges(ctx)
		return err
	}
	err := rm.replay(ctx, operation)
	return gauges, err
}

func (rm *ErrorStorageMiddleware) GetAllCounters(ctx context.Context) (map[string]int64, error) {
	var counters map[string]int64
	operation := func() error {
		var err error
		counters, err = rm.next.GetAllCounters(ctx)
		return err
	}
	err := rm.replay(ctx, operation)
	return counters, err
}

func (rm *ErrorStorageMiddleware) SaveAll(ctx context.Context, metrics []models.Metrics) error {
	operation := func() error {
		return rm.next.SaveAll(ctx, metrics)
	}
	return rm.replay(ctx, operation)
}

func (rm *ErrorStorageMiddleware) Ping(ctx context.Context) error {
	operation := func() error {
		return rm.next.Ping(ctx)
	}
	return rm.replay(ctx, operation)
}
