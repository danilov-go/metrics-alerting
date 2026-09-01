// Package db реализует хранилище данных в базе PostgreSQL.
package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/danilov-go/metrics-alerting.git/internal/models"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type storageDB struct {
	db *sql.DB
}

// NewStorageDB создает новый экземпляр storageDB.
func NewStorageDB(sql *sql.DB) *storageDB {
	return &storageDB{
		db: sql,
	}
}

// InitDB инициализирует подключение к базе данных PostgreSQL по переданной строке соединения.
func InitDB(ps string) (*storageDB, error) {
	db, err := sql.Open("pgx", ps)
	if err != nil {
		return nil, err
	}
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		if closeErr := db.Close(); closeErr != nil {
			return nil, errors.Join(err, closeErr)
		}
		return nil, err
	}
	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres",
		driver,
	)
	if err != nil {
		if closeErr := db.Close(); closeErr != nil {
			return nil, errors.Join(err, closeErr)
		}
		return nil, err
	}
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		if closeErr := db.Close(); closeErr != nil {
			return nil, errors.Join(err, closeErr)
		}
		return nil, err
	}
	return &storageDB{
		db: db,
	}, nil
}

// SaveCounters сохраняет или обновляет метрику типа "counter".
func (d *storageDB) SaveCounters(ctx context.Context, name string, delta int64) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	query := `
		INSERT INTO metrics (id, mtype, delta) VALUES ($1, 'counter', $2)
		ON CONFLICT (id) 
		DO UPDATE SET delta = metrics.delta + EXCLUDED.delta
		`
	_, err := d.db.ExecContext(ctx, query, name, delta)
	if err != nil {
		return err
	}
	return nil
}

// SaveGauges сохраняет или перезаписывает метрику типа "gauge".
func (d *storageDB) SaveGauges(ctx context.Context, name string, value float64) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	query := `
		INSERT INTO metrics (id, mtype, value) VALUES ($1, 'gauge', $2)
		ON CONFLICT (id) 
		DO UPDATE SET value = EXCLUDED.value
		`
	_, err := d.db.ExecContext(ctx, query, name, value)
	if err != nil {
		return err
	}
	return nil
}

// GetGauges возвращает значение метрики типа "gauge" по её названию.
func (d *storageDB) GetGauges(ctx context.Context, name string) (float64, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	query := `SELECT value FROM metrics WHERE mtype = 'gauge' AND id = $1`
	var value sql.NullFloat64
	row := d.db.QueryRowContext(ctx, query, name)
	err := row.Scan(&value)
	if err != nil {
		return 0, err
	}
	if !value.Valid {
		return 0, err
	}
	return value.Float64, nil
}

// GetCounters возвращает значение метрики типа "counter" по её названию.
func (d *storageDB) GetCounters(ctx context.Context, name string) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	query := `SELECT delta FROM metrics WHERE mtype = 'counter' AND id = $1`
	var delta sql.NullInt64
	row := d.db.QueryRowContext(ctx, query, name)
	err := row.Scan(&delta)
	if err != nil {
		return 0, err
	}
	if !delta.Valid {
		return 0, err
	}
	return delta.Int64, nil
}

// GetAllGauges возвращает map всех хранящихся метрик типа "gauge".
func (d *storageDB) GetAllGauges(ctx context.Context) (map[string]float64, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	query := `SELECT id, value FROM metrics WHERE mtype = 'gauge'`
	gauges := make(map[string]float64)
	rows, err := d.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()
	for rows.Next() {
		var id string
		var value sql.NullFloat64
		err = rows.Scan(&id, &value)
		if err != nil {
			return nil, err
		}
		if value.Valid {
			gauges[id] = value.Float64
		}
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return gauges, nil
}

// GetAllCounters возвращает map всех хранящихся метрик типа "counter".
func (d *storageDB) GetAllCounters(ctx context.Context) (map[string]int64, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	query := `SELECT id, delta FROM metrics WHERE mtype = 'counter'`
	counters := make(map[string]int64)
	rows, err := d.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()
	for rows.Next() {
		var id string
		var delta sql.NullInt64
		err = rows.Scan(&id, &delta)
		if err != nil {
			return nil, err
		}
		if delta.Valid {
			counters[id] = delta.Int64
		}
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return counters, nil
}

// SaveAll выполняет пакетное сохранение метрик в рамках одной транзакции.
func (d *storageDB) SaveAll(ctx context.Context, metrics []models.Metrics) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()
	stmtCounter, err := tx.PrepareContext(ctx, `
		INSERT INTO metrics (id, mtype, delta) VALUES ($1, 'counter', $2)
		ON CONFLICT (id) 
		DO UPDATE SET delta = metrics.delta + EXCLUDED.delta
	`)
	if err != nil {
		return err
	}
	defer func() {
		_ = stmtCounter.Close()
	}()
	stmtGauge, err := tx.PrepareContext(ctx, `
		INSERT INTO metrics (id, mtype, value) VALUES ($1, 'gauge', $2)
		ON CONFLICT (id) 
		DO UPDATE SET value = EXCLUDED.value
		`)
	if err != nil {
		return err
	}
	defer func() {
		_ = stmtCounter.Close()
	}()
	for _, m := range metrics {
		switch m.MType {
		case models.Counter:
			if m.Delta == nil {
				return fmt.Errorf("delta равно nil")
			}
			_, err = stmtCounter.ExecContext(ctx, m.ID, *m.Delta)
			if err != nil {
				return err
			}
		case models.Gauge:
			if m.Value == nil {
				return fmt.Errorf("value равно nil")
			}
			_, err = stmtGauge.ExecContext(ctx, m.ID, *m.Value)
			if err != nil {
				return err
			}
		}
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

// Ping проверяет доступность хранилища.
func (d *storageDB) Ping(ctx context.Context) error {
	err := d.db.PingContext(ctx)
	if err != nil {
		return err
	}
	return nil
}
