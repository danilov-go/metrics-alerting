package db

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/danilov-go/metrics-alerting.git/internal/models"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type log interface {
	Errorw(msg string, keysAndValues ...any)
}

type storageDB struct {
	DB     *sql.DB
	Logger log
}

func InitDB(ps string, l log) (*storageDB, error) {
	db, err := sql.Open("pgx", ps)
	if err != nil {
		return nil, err
	}
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		db.Close()
		return nil, err
	}
	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres",
		driver,
	)
	if err != nil {
		db.Close()
		return nil, err
	}
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		db.Close()
		return nil, err
	}
	return &storageDB{
		DB:     db,
		Logger: l,
	}, nil
}

func (d *storageDB) SaveCounters(ctx context.Context, name string, delta int64) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	query := `
		INSERT INTO metrics (id, mtype, delta) VALUES ($1, 'counter', $2)
		ON CONFLICT (id) 
		DO UPDATE SET delta = metrics.delta + EXCLUDED.delta
		`
	_, err := d.DB.ExecContext(ctx, query, name, delta)
	if err != nil {
		d.Logger.Errorw("ошибка сохранения counter", err)
		return err
	}
	return nil
}

func (d *storageDB) SaveGauges(ctx context.Context, name string, value float64) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	query := `
		INSERT INTO metrics (id, mtype, value) VALUES ($1, 'gauge', $2)
		ON CONFLICT (id) 
		DO UPDATE SET value = EXCLUDED.value
		`
	_, err := d.DB.ExecContext(ctx, query, name, value)
	if err != nil {
		d.Logger.Errorw("ошибка сохранения gauge", err)
		return err
	}
	return nil
}

func (d *storageDB) GetGauges(ctx context.Context, name string) (float64, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	query := `SELECT value FROM metrics WHERE mtype = 'gauge' AND id = $1`
	var value sql.NullFloat64
	row := d.DB.QueryRowContext(ctx, query, name)
	err := row.Scan(&value)
	if err != nil {
		return 0, err
	}
	if !value.Valid {
		return 0, err
	}
	return value.Float64, nil
}

func (d *storageDB) GetCounters(ctx context.Context, name string) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	query := `SELECT delta FROM metrics WHERE mtype = 'counter' AND id = $1`
	var delta sql.NullInt64
	row := d.DB.QueryRowContext(ctx, query, name)
	err := row.Scan(&delta)
	if err != nil {
		return 0, err
	}
	if !delta.Valid {
		return 0, err
	}
	return delta.Int64, nil
}

func (d *storageDB) GetAllGauges(ctx context.Context) (map[string]float64, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	query := `SELECT id, value FROM metrics WHERE mtype = 'gauge'`
	gauges := make(map[string]float64)
	rows, err := d.DB.QueryContext(ctx, query)
	if err != nil {
		d.Logger.Errorw("ошибка выполнения запроса", "err", err)
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id string
		var value sql.NullFloat64
		err := rows.Scan(&id, &value)
		if err != nil {
			d.Logger.Errorw("ошибка сканирования", "err", err)
			return nil, err
		}
		if value.Valid {
			gauges[id] = value.Float64
		}
	}

	err = rows.Err()
	if err != nil {
		d.Logger.Errorw("ошибка чтения строк", "err", err)
		return nil, err
	}
	return gauges, nil
}

func (d *storageDB) GetAllCounters(ctx context.Context) (map[string]int64, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	query := `SELECT id, delta FROM metrics WHERE mtype = 'counter'`
	counters := make(map[string]int64)
	rows, err := d.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id string
		var delta sql.NullInt64
		err := rows.Scan(&id, &delta)
		if err != nil {
			d.Logger.Errorw("ошибка сканирования", "err", err)
			return nil, err
		}
		if delta.Valid {
			counters[id] = delta.Int64
		}
	}
	err = rows.Err()
	if err != nil {
		d.Logger.Errorw("ошибка чтения строк", "err", err)
		return nil, err
	}
	return counters, nil
}

func (d *storageDB) SaveAll(ctx context.Context, metrics []models.Metrics) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	tx, err := d.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	stmtCounter, err := tx.PrepareContext(ctx, `
		INSERT INTO metrics (id, mtype, delta) VALUES ($1, 'counter', $2)
		ON CONFLICT (id) 
		DO UPDATE SET delta = metrics.delta + EXCLUDED.delta
	`)
	if err != nil {
		return err
	}
	defer stmtCounter.Close()
	stmtGauge, err := tx.PrepareContext(ctx, `
		INSERT INTO metrics (id, mtype, value) VALUES ($1, 'gauge', $2)
		ON CONFLICT (id) 
		DO UPDATE SET value = EXCLUDED.value
		`)
	if err != nil {
		return err
	}
	defer stmtGauge.Close()
	for _, m := range metrics {
		switch m.MType {
		case models.Counter:
			if m.Delta == nil {
				d.Logger.Errorw("delta равна nil")
				continue
			}
			_, err := stmtCounter.ExecContext(ctx, m.ID, *m.Delta)
			if err != nil {
				d.Logger.Errorw("ошибка сохранения counter", err)
				return err
			}
		case models.Gauge:
			if m.Value == nil {
				d.Logger.Errorw("value равно nil")
				continue
			}
			_, err := stmtGauge.ExecContext(ctx, m.ID, *m.Value)
			if err != nil {
				d.Logger.Errorw("ошибка сохранения gauge", err)
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
