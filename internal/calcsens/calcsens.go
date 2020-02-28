package calcsens

import (
	"fmt"
	"github.com/fpawel/sensel/internal/data"
	"math"
	"time"
)

type measurement struct {
	Name     string
	Gas      int
	Duration time.Duration
	Calc     calcSamplesFunc
}

type calcSamplesFunc = func(*measureData) (float64, bool)

type measureData struct {
	U, I, Q, T, C float64

	dataMeasure data.Measurement
	prodType    productType
	calculated  []SampleCalc
	place       int
	err         error
}

func (x measureData) Measure(name string) prevMeasureData {
	m, foundMeasurement := x.prodType.measurementByName(name)
	if !foundMeasurement {
		x.err = fmt.Errorf("измерение %q не задано для исполнения %q", name, x.dataMeasure.ProductType)
		return nanMeasureData
	}
	for _, r := range x.calculated {
		if r.Name == name {
			if len(x.dataMeasure.Pgs) <= m.Gas {
				x.err = fmt.Errorf("нет значения ПГС%d", m.Gas)
				return nanMeasureData
			}
			return prevMeasureData{
				Value: r.Calc[x.place].Float,
				U:     r.Productions[x.place].Value,
				I:     r.Current,
				Q:     r.Consumption,
				T:     r.Temperature,
				C:     x.dataMeasure.Pgs[m.Gas],
			}
		}
	}
	x.err = fmt.Errorf("измерение %q не было выполнено", name)
	return nanMeasureData
}

type prevMeasureData struct {
	Value, U, I, Q, T, C float64
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
