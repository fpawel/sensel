package data

import (
	"database/sql"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDB(t *testing.T) {
	db := newTestDB(t)

	Assert := assert.New(t)

	_, err := db.Exec(querySchema)
	Assert.NoError(err)

	m := simpleMeasurement()

	Assert.NoError(SaveMeasurement(db, &m))
	Assert.Equal(int64(1), m.MeasurementID)

	m.Pgs = []float64{6, 7, 8, 9}
	m.Name = "000000"
	m.ProductType = "111111"
	Assert.NoError(SaveMeasurement(db, &m))

	var m2 Measurement
	m2.MeasurementID = m.MeasurementID
	Assert.NoError(GetMeasurement(db, &m2))
	Assert.Equal(m, m2)
}

func newTestDB(t *testing.T) *sqlx.DB {
	conn, err := sql.Open("sqlite3", ":memory:")
	assert.NoError(t, err)
	return sqlx.NewDb(conn, "sqlite3")
}
func simpleMeasurement() Measurement {
	return Measurement{
		MeasurementInfo: MeasurementInfo{
			MeasurementID: 0,
			Name:          "abc",
			ProductType:   "def",
		},
		MeasurementData: MeasurementData{
			Pgs: []float64{1, 2, 3, 4},
			Samples: []Sample{
				{
					Name:        "R0A",
					Productions: randProductions(),
				},
			},
		},
	}
}
