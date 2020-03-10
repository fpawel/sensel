package data

import (
	"database/sql"
	"github.com/fpawel/sensel/internal/pkg/cmpreport"
	"github.com/google/go-cmp/cmp"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestCRUD(t *testing.T) {
	db := newTestDB(t)

	Assert := assert.New(t)

	_, err := db.Exec(querySchema)
	Assert.NoError(err)

	m := simpleMeasurement1(time.Now())

	Assert.NoError(SaveMeasurement(db, &m))
	Assert.Equal(int64(1), m.MeasurementID)

	m.Pgs = []float64{6, 7, 8, 9}
	m.Name = "000000"
	m.ProductType = "111111"
	m.Samples = make(Samples, 10)
	RandSamples(m.Samples)
	Assert.NoError(SaveMeasurement(db, &m))

	var m2 Measurement
	m2.MeasurementID = m.MeasurementID
	Assert.NoError(GetMeasurement(db, &m2))
	cmpreport.AssertEqual(t, m, m2)

	m.MeasurementID = 0
	m.CreatedAt = m.CreatedAt.Add(time.Hour)
	RandSamples(m.Samples)
	Assert.NoError(SaveMeasurement(db, &m))

	m2.MeasurementID = m.MeasurementID
	Assert.NoError(GetMeasurement(db, &m2))
	cmpreport.AssertEqual(t, m, m2)

}

func newTestDB(t *testing.T) *sqlx.DB {
	conn, err := sql.Open("sqlite3", ":memory:")
	assert.NoError(t, err)
	return sqlx.NewDb(conn, "sqlite3")
}

func simpleMeasurement1(createdAt time.Time) Measurement {
	return Measurement{
		MeasurementInfo: MeasurementInfo{
			MeasurementID: 0,
			Name:          "abc",
			ProductType:   "СГГ-1",
			CreatedAt:     createdAt,
		},
		MeasurementData: MeasurementData{
			Pgs: []float64{1, 2, 3, 4},
			Samples: []Sample{
				newSample("R0A"),
				newSample("Uр"),
				newSample("T20A"),
				newSample("U20A"),
			},
		},
	}
}

func simpleMeasurement2(createdAt time.Time) Measurement {
	return Measurement{
		MeasurementInfo: MeasurementInfo{
			MeasurementID: 0,
			Name:          "def",
			ProductType:   "СТМ-10 СКДМ",
			CreatedAt:     createdAt,
		},
		MeasurementData: MeasurementData{
			Pgs: []float64{1, 2, 3, 4},
			Samples: []Sample{
				newSample("X5"),
				newSample("X6"),
			},
		},
	}
}

func newSample(name string) Sample {
	return Sample{
		Name:        name,
		Productions: randProductions(),
	}
}

type timeTime time.Time

func (x timeTime) Equal(y timeTime) bool {
	return time.Time(x).Equal(time.Time(y))
}

// This transformer converts otherString to myString, allowing Equal to use
// other Options to determine equality.
var transformTime = cmp.Transformer("", func(in time.Time) timeTime {
	return timeTime(in)
})
