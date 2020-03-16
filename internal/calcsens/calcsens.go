package calcsens

import (
	"fmt"
	"github.com/fpawel/sensel/internal/data"
	"math"
	"time"
)

type column struct {
	Name     string
	Gas      int
	Duration time.Duration
	Calc     calcSamplesFunc
}

type calcSamplesFunc = func(*measureData) (float64, bool)

type measureData struct {
	U, I, Q, T, C float64
	place         int
	d             data.Measurement
	prodType      productType
	calculated    []ColumnCalculated
	err           error
}

func (x column) Column() Column {
	return Column{
		Name:     x.Name,
		Gas:      x.Gas,
		Duration: x.Duration,
	}
}

func (x measureData) Measure(name string) prevMeasureData {
	m, foundMeasurement := x.prodType.measurementByName(name)
	if !foundMeasurement {
		x.err = fmt.Errorf("измерение %q не задано для исполнения %q", name, x.d.ProductType)
		return nanMeasureData
	}
	for _, r := range x.calculated {
		if r.Name == name {
			if len(x.d.Pgs) <= m.Gas {
				x.err = fmt.Errorf("нет значения ПГС%d для расчёта %q", m.Gas, name)
				return nanMeasureData
			}
			pd := nanMeasureData
			pd.Value = r.Values[x.place].Float
			pd.Pgs = x.d.Pgs[m.Gas]
			for _, smp := range x.d.Samples {
				if smp.Name == name {
					pd.U = smp.Productions[x.place].Value
					pd.I = smp.CurrentBar
					pd.Q = smp.Consumption
					pd.T = smp.Temperature
				}
			}
			return pd
		}
	}
	x.err = fmt.Errorf("измерение %q не было выполнено", name)
	return nanMeasureData
}

type prevMeasureData struct {
	Value, U, I, Q, T, Pgs float64
}

var (
	nanMeasureData = prevMeasureData{
		Value: math.NaN(),
		U:     math.NaN(),
		I:     math.NaN(),
		Q:     math.NaN(),
		T:     math.NaN(),
	}
)
