package data

import (
	"database/sql"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"time"
)

//go:generate go run github.com/fpawel/gotools/cmd/sqlstr/...

func Open(filename string) (*sqlx.DB, error) {
	db, err := open(filename)
	if err != nil {
		return nil, err
	}
	if _, err := db.Exec(SQLCreate); err != nil {
		return nil, err
	}
	return db, nil
}

func NewMeasurement(db *sqlx.DB) (int64, error) {
	r, err := db.Exec(`INSERT INTO measurement (created_at)  VALUES (?)`, time.Now())
	if err != nil {
		return 0, err
	}
	return getNewInsertedID(r)
}

func open(fileName string) (*sqlx.DB, error) {
	conn, err := sql.Open("sqlite3", fileName)
	if err != nil {
		return nil, err
	}
	conn.SetMaxIdleConns(1)
	conn.SetMaxOpenConns(1)
	conn.SetConnMaxLifetime(0)
	return sqlx.NewDb(conn, "sqlite3"), nil
}

func getNewInsertedID(r sql.Result) (int64, error) {
	n, err := r.RowsAffected()
	if err != nil {
		return 0, err
	}
	if n != 1 {
		return 0, fmt.Errorf("excpected 1 rows inserted, got %d", n)
	}
	id, err := r.LastInsertId()
	if err != nil {
		return 0, err
	}
	if id <= 0 {
		return 0, merry.New("was not inserted")
	}
	return id, nil
}
