package db

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type storageDB struct {
	DB *sql.DB
}

const schema = `
CREATE TABLE IF NOT EXISTS metrics (
    id VARCHAR(255) PRIMARY KEY,
    mtype VARCHAR(255) NOT NULL,
    delta BIGINT,
    value DOUBLE PRECISION
);

CREATE INDEX IF NOT EXISTS idx_metrics_mtype ON metrics(mtype);
`

func InitDB(ps string) (*storageDB, error) {
	db, err := sql.Open("pgx", ps)
	if err != nil {
		return nil, err
	}
	_, err = db.Exec(schema)
	if err != nil {
		return nil, err
	}
	return &storageDB{
		DB: db,
	}, nil
}

func (d *storageDB) SaveCounters(name string, delta int64) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	deltaOld, valid := d.GetCounters(name)
	if !valid {
		query := `INSERT INTO metrics (id, mtype, delta) VALUES ($1, 'counter', $2)`
		_, _ = d.DB.ExecContext(ctx, query, name, delta)
	} else {
		query := `UPDATE metrics SET delta = $1 WHERE id = $2 AND mtype = 'counter'`
		_, _ = d.DB.ExecContext(ctx, query, delta+deltaOld, name)
	}
}

func (d *storageDB) SaveGauges(name string, value float64) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, valid := d.GetGauges(name)
	if !valid {
		query := `INSERT INTO metrics (id, mtype, value) VALUES ($1, 'gauge', $2)`
		_, _ = d.DB.ExecContext(ctx, query, name, value)
	} else {
		query := `UPDATE metrics SET value = $1 WHERE id = $2 AND mtype = 'gauge'`
		_, _ = d.DB.ExecContext(ctx, query, value, name)
	}
}

func (d *storageDB) GetGauges(name string) (float64, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	query := `SELECT value FROM metrics WHERE mtype = 'gauge' AND id = $1`
	var value sql.NullFloat64
	row := d.DB.QueryRowContext(ctx, query, name)
	err := row.Scan(&value)
	if err != nil || !value.Valid {
		return 0, false
	}
	return value.Float64, true
}

func (d *storageDB) GetCounters(name string) (int64, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	query := `SELECT delta FROM metrics WHERE mtype = 'counter' AND id = $1`
	var delta sql.NullInt64
	row := d.DB.QueryRowContext(ctx, query, name)
	err := row.Scan(&delta)
	if err != nil || !delta.Valid {
		return 0, false
	}
	return delta.Int64, true
}

func (d *storageDB) GetAllGauges() map[string]float64 {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	query := `SELECT id, value FROM metrics WHERE mtype = 'gauge'`
	gauges := make(map[string]float64)
	rows, err := d.DB.QueryContext(ctx, query)
	if err != nil {
		return nil
	}
	defer rows.Close()
	for rows.Next() {
		var id string
		var value float64
		err := rows.Scan(&id, &value)
		if err != nil {
			return nil
		}
		gauges[id] = value
	}
	err = rows.Err()
	if err != nil {
		return nil
	}
	return gauges
}

func (d *storageDB) GetAllCounters() map[string]int64 {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	query := `SELECT id, delta FROM metrics WHERE mtype = 'counter'`
	counters := make(map[string]int64)
	rows, err := d.DB.QueryContext(ctx, query)
	if err != nil {
		return nil
	}
	defer rows.Close()
	for rows.Next() {
		var id string
		var delta int64
		err := rows.Scan(&id, &delta)
		if err != nil {
			return nil
		}
		counters[id] = delta
	}
	if err = rows.Err(); err != nil {
		return nil
	}
	return counters
}

func (d *storageDB) SaveFile(path string) error {
	return nil
}
