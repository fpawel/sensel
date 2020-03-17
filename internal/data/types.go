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
	Samples []Sample
}

type MeasurementInfo struct {
	MeasurementID int64     `db:"measurement_id"`
	CreatedAt     time.Time `db:"created_at"`
	Name          string    `db:"name"`
	Device        string    `db:"device"`
	Kind          string    `db:"kind"`
}

type Sample struct {
	CreatedAt   time.Time
	Gas         int
	Q           float64
	T           float64
	U           float64
	I           float64
	Productions [16]Production
}

type Production struct {
	Place int     `db:"int"`
	Value float64 `db:"value"`
	Break bool    `db:"break"`
}
