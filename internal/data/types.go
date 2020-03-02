package data

import (
	"time"
)

type Measurement struct {
	MeasurementID int64     `db:"measurement_id"`
	CreatedAt     time.Time `db:"created_at"`
	Name          string    `db:"name"`
	ProductType   string    `db:"product_type"`
	Pgs           []float64 `db:"-"`
	Samples       Samples   `db:"-"`
}

type Sample struct {
	SampleID    int64        `db:"sample_id"`
	CreatedAt   time.Time    `db:"created_at"`
	Name        string       `db:"name"`
	Gas         int          `db:"gas"`
	Consumption float64      `db:"consumption"`
	Temperature float64      `db:"temperature"`
	Current     float64      `db:"current"`
	Productions []Production `db:"-"`
	Logs        []SampleLog  `db:"-"`
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
