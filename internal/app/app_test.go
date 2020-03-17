package app

import (
	"fmt"
	"github.com/fpawel/sensel/internal/data"
	"github.com/fpawel/sensel/internal/pkg/cmpreport"
	"github.com/stretchr/testify/require"
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
	db, err := data.Open(dbFilename)
	require.NoError(t, err)

	defer func() {
		require.NoError(t, db.Close())
	}()

	_, err = db.Exec(`DELETE FROM measurement WHERE TRUE`)
	require.NoError(t, err)

	tm := time.Now()
	for i := 0; i < 100; i++ {
		samples := make([]data.Sample, 10)
		data.RandSamples(samples)
		for i := range samples {
			samples[i].CreatedAt = tm
			tm = tm.Add(-time.Hour)
		}
		m := data.Measurement{
			MeasurementInfo: data.MeasurementInfo{
				CreatedAt: tm,
				Device:    "СТМ-30",
				Kind:      "сравнительный",
			},
			MeasurementData: data.MeasurementData{
				Pgs:     []float64{1, 2, 3, 4, 5},
				Samples: samples,
			},
		}
		tm = tm.Add(-time.Hour)
		require.NoError(t, data.SaveMeasurement(db, &m), fmt.Sprintf("%+v", m))

		m.Name = fmt.Sprintf("%d", m.MeasurementID)
		require.NoError(t, data.SaveMeasurement(db, &m), fmt.Sprintf("%+v", m))

		var m1 data.Measurement
		m1.MeasurementID = m.MeasurementID
		require.NoError(t, data.GetMeasurement(db, &m1))
		cmpreport.AssertEqual(t, m, m1)
	}
}
