package data

import (
	"time"
)

type Measurement struct {
	MeasurementInfo
	MeasurementData
}

type MeasurementData struct {
	Pgs     []float64
	Samples Samples
}

type MeasurementInfo struct {
	MeasurementID int64     `db:"measurement_id"`
	CreatedAt     time.Time `db:"created_at"`
	Name          string    `db:"name"`
	ProductType   string    `db:"product_type"`
}

type Sample struct {
	CreatedAt   time.Time
	Name        string
	Gas         int
	Consumption float64
	Temperature float64
	Current     float64
	Productions [16]Production
}

type Production struct {
	Place int     `db:"int"`
	Value float64 `db:"value"`
	Break bool    `db:"break"`
}

type SampleLog struct {
	CreatedAt time.Time `db:"created_at"`
	Ok        bool      `db:"ok"`
	Message   string    `db:"message"`
}

type Samples []Sample

func (x Samples) GetSampleByName(name string) (Sample, bool) {
	for _, smp := range x {
		if smp.Name == name {
			return smp, true
		}
	}
	return Sample{}, false

}
