package db

import (
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type storageDB struct {
	DB *sql.DB
}

func InitDB(ps string) (*storageDB, error) {
	db, err := sql.Open("pgx", ps)
	if err != nil {
		return nil, err
	}
	return &storageDB{
		DB: db,
	}, nil
}

func (d *storageDB) SaveCounters(name string, value int64) {
}
func (d *storageDB) SaveGauges(name string, value float64) {
}
func (d *storageDB) GetGauges(name string) (float64, bool) {
	return 0, true
}
func (d *storageDB) GetCounters(name string) (int64, bool) {
	return 0, true
}
func (d *storageDB) GetAllGauges() map[string]float64 {
	return nil
}
func (d *storageDB) GetAllCounters() map[string]int64 {
	return nil
}
func (d *storageDB) SaveFile(path string) error {
	return nil
}
