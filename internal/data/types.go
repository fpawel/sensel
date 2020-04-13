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
	Tm  time.Time
	Gas int
	Q   float64
	T   float64
	Ub  float64
	I   float64
	U   [16]float64
	Br  [16]bool
}

func (m Measurement) I() (xs []float64) {
	for _, smp := range m.Samples {
		xs = append(xs, smp.I)
	}
	return
}

func (m Measurement) T() (xs []float64) {
	for _, smp := range m.Samples {
		xs = append(xs, smp.T)
	}
	return
}

func (m Measurement) U(n int) (xs []float64) {
	for _, smp := range m.Samples {
		xs = append(xs, smp.U[n])
	}
	return
}

func (m Measurement) Br(n int) bool {
	for _, smp := range m.Samples {
		if smp.Br[n] {
			return true
		}
	}
	return false
}
