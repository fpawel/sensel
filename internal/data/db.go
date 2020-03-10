package data

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

func Open(filename string) (*sqlx.DB, error) {

	db, err := open(filename)
	if err != nil {
		return nil, err
	}
	if _, err := db.Exec(querySchema); err != nil {
		return nil, err
	}
	return db, nil
}

func GetLastMeasurement(db *sqlx.DB, m *Measurement) error {
	var x struct {
		MeasurementInfo
		Data []byte `db:"data"`
	}
	const query = `SELECT * FROM measurement ORDER BY created_at DESC LIMIT 1`
	if err := db.Get(&x, query); err != nil {
		return err
	}
	m.MeasurementInfo = x.MeasurementInfo
	if err := json.Unmarshal(x.Data, &m.MeasurementData); err != nil {
		return err
	}
	return nil
}

func GetMeasurement(db *sqlx.DB, m *Measurement) error {
	var x struct {
		MeasurementInfo
		Data []byte `db:"data"`
	}
	const query = `SELECT * FROM measurement WHERE measurement_id=?`

	if err := db.Get(&x, query, m.MeasurementID); err != nil {
		return err
	}
	m.MeasurementInfo = x.MeasurementInfo
	if err := json.Unmarshal(x.Data, &m.MeasurementData); err != nil {
		return err
	}
	return nil
}

func SaveMeasurement(db *sqlx.DB, m *Measurement) error {
	const query = `
INSERT INTO measurement (measurement_id, created_at, product_type, name, data)  
VALUES (?, ?, ?, ?, ?)
ON CONFLICT (measurement_id) DO UPDATE SET product_type = ?, name=?, data=?
`
	b, err := json.Marshal(m.MeasurementData)
	if err != nil {
		return err
	}

	var measurementID interface{} = nil
	if m.MeasurementID > 0 {
		measurementID = m.MeasurementID
	}

	r, err := db.Exec(query, measurementID, m.CreatedAt,
		m.ProductType, m.Name, b,
		m.ProductType, m.Name, b)
	if err != nil {
		return err
	}

	n, err := r.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		return fmt.Errorf("excpected 1 rows inserted, got %d", n)
	}

	if id, err := r.LastInsertId(); err == nil {
		m.MeasurementID = id
	}
	return nil
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

const querySchema = `
PRAGMA foreign_keys = ON;
PRAGMA encoding = 'UTF-8';

CREATE TABLE IF NOT EXISTS measurement
(
    measurement_id INTEGER PRIMARY KEY,
    created_at     TIMESTAMP NOT NULL,
    product_type   TEXT      NOT NULL DEFAULT '',
    name           TEXT      NOT NULL DEFAULT '',
    data           BLOB      NOT NULL
);
CREATE INDEX IF NOT EXISTS index_measurement_created_at ON measurement(created_at);
`
