package data

import (
	"database/sql"
	"fmt"
	"github.com/fpawel/sensel/internal/pkg/cmpreport"
	"github.com/google/go-cmp/cmp"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCreateDB(t *testing.T) {
	exeDir := filepath.Join(os.Getenv("GOPATH"), "src", "github.com", "fpawel", "sensel", "build")
	//prodTypes, err := calcsens.NewProductTypes(filepath.Join(exeDir, "lua", "sensel.lua"))
	//require.NoError(t, err)

	// соединение с базой данных
	dbFilename := filepath.Join(exeDir, "sensel.sqlite")
	db, err := Open(dbFilename)
	require.NoError(t, err)

	defer func() {
		require.NoError(t, db.Close())
	}()

	_, err = db.Exec(`DELETE FROM measurement WHERE TRUE`)
	require.NoError(t, err)

	tm := time.Now()
	for i := 0; i < 100; i++ {
		samples := make([]Sample, 10)
		randSamples(samples)
		for i := range samples {
			samples[i].Tm = tm
			tm = tm.Add(-time.Hour)
		}
		m := Measurement{
			MeasurementInfo: MeasurementInfo{
				CreatedAt: tm,
				Device:    "СГГ-1",
				Kind:      "измерительный",
			},
			MeasurementData: MeasurementData{
				Pgs:     []float64{rand3(), rand3(), rand3(), rand3(), rand3()},
				Samples: samples,
			},
		}
		tm = tm.Add(-time.Hour * 24)
		require.NoError(t, SaveMeasurement(db, &m), fmt.Sprintf("%+v", m))

		m.Name = fmt.Sprintf("%v", m.CreatedAt)
		require.NoError(t, SaveMeasurement(db, &m), fmt.Sprintf("%+v", m))

		var m1 Measurement
		m1.MeasurementID = m.MeasurementID
		require.NoError(t, GetMeasurement(db, &m1))
		cmpreport.AssertEqual(t, m, m1)
	}
}

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
	randSamples(m.Samples)
	Assert.NoError(SaveMeasurement(db, &m))

	var m2 Measurement
	m2.MeasurementID = m.MeasurementID
	Assert.NoError(GetMeasurement(db, &m2))
	cmpreport.AssertEqual(t, m, m2)

	m.MeasurementID = 0
	m.CreatedAt = m.CreatedAt.Add(time.Hour)
	randSamples(m.Samples)
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
	samples := make([]Sample, 10)
	randSamples(samples)
	return Measurement{
		MeasurementInfo: MeasurementInfo{
			MeasurementID: 0,
			Name:          "abc",
			Device:        "СГГ-1",
			Kind:          "измерительный",
			CreatedAt:     createdAt,
		},
		MeasurementData: MeasurementData{
			Pgs:     []float64{1, 2, 3, 4},
			Samples: samples,
		},
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

func randSamples(xs []Sample) {
	for i := range xs {
		xs[i].Gas = i % 4
		xs[i].T = rand3()
		xs[i].I = rand3()
		xs[i].Q = rand3()
		for j := range xs[i].U {
			xs[i].U[j] = rand3()
			xs[i].Br[j] = rand.Float64() < 0.3
		}
	}
}

func rand3() float64 {
	return math.Round(rand.Float64()*1000) / 1000
}

func init() {
	rand.NewSource(time.Now().UnixNano())
}
