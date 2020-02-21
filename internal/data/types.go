package data

import (
	"time"
)

type Measurement struct {
	MeasurementID int64     `db:"measurement_id"`
	CreatedAt     time.Time `db:"created_at"`
	Name          string    `db:"name"`
	Pgs           []float64 `db:"-"`
	Samples       []Sample  `db:"-"`
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
	Logs        []Log        `db:"-"`
}

type Production struct {
	Value        float64 `db:"value"`
	BreakCircuit bool    `db:"break_circuit"`
	Ok           bool    `db:"ok"`
	Message      string  `db:"message"`
}

type Log struct {
	CreatedAt time.Time `db:"created_at"`
	Ok        bool      `db:"ok"`
	Message   string    `db:"message"`
}
