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
	m.Device = "111111"
	m.Kind = "222222"
	m.Samples = make([]Sample, 10)
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
			Device:        "СГГ-1",
			Kind:          "измерительный",
			CreatedAt:     createdAt,
		},
		MeasurementData: MeasurementData{
			Pgs: []float64{1, 2, 3, 4},
			Samples: []Sample{
				newSample(),
				newSample(),
				newSample(),
				newSample(),
			},
		},
	}
}

func simpleMeasurement2(createdAt time.Time) Measurement {
	return Measurement{
		MeasurementInfo: MeasurementInfo{
			MeasurementID: 0,
			Name:          "def",
			Device:        "СТМ-10 СКДМ",
			Kind:          "сравнительный",
			CreatedAt:     createdAt,
		},
		MeasurementData: MeasurementData{
			Pgs: []float64{1, 2, 3, 4},
			Samples: []Sample{
				newSample(),
				newSample(),
			},
		},
	}
}

func newSample() Sample {
	return Sample{
		Productions: randProductions(),
		CreatedAt:   time.Now(),
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
