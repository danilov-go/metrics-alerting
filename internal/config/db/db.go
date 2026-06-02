package db

import (
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"
)

var DB *sql.DB

func InitDB(ps string) error {
	db, err := sql.Open("pgx", ps)
	if err != nil {
		return err
	}
	DB = db
	return nil
}
