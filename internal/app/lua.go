package app

import (
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/sensel/internal/data"
	"github.com/fpawel/sensel/internal/pkg/must"
	lua "github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"
	"math"
	"os"
	"path/filepath"
	"time"
)

type ProductType struct {
	Measurements []Measurement
}

func (x ProductType) MeasurementByName(name string) (Measurement, int) {
	for n, x := range x.Measurements {
		if x.Name == name {
			return x, n
		}
	}
	return Measurement{}, -1
}

type Measurement struct {
	Name     string
	Gas      int
	Duration time.Duration
	Calc     CalcSamplesFunc
}

type CalcSamplesFunc = func(*measureData) (float64, bool)

type measureData struct {
	U, I, Q, T, C float64

	prodType     ProductType
	prodTypeName string
	calc         []SampleCalc
	place        int
	err          error
	pgs          []float64
}

func (x measureData) Measure(name string) prevMeasureData {
	m, n := x.prodType.MeasurementByName(name)
	if n == -1 {
		x.err = fmt.Errorf("измерение %q не задано для исполнения %q", name, x.prodTypeName)
		return nanMeasureData
	}
	for _, r := range x.calc {
		if r.Name == name {
			if len(x.pgs) <= m.Gas {
				x.err = fmt.Errorf("нет значения ПГС%d", m.Gas)
				return nanMeasureData
			}
			return prevMeasureData{
				Value: r.Calc[x.place].V,
				U:     r.Productions[x.place].Value,
				I:     r.Current,
				Q:     r.Consumption,
				T:     r.Temperature,
				C:     x.pgs[m.Gas],
			}
		}
	}
	x.err = fmt.Errorf("измерение %q не было выполнено", name)
	return nanMeasureData
}

type prevMeasureData struct {
	Value, U, I, Q, T, C float64
}

type SampleCalc struct {
	data.Sample
	Calc []Float64Ok
}

type Float64Ok struct {
	V  float64
	Ok bool
}

func CalcSamples(measurement data.Measurement) ([]SampleCalc, error) {
	prodType, okProdType := ProductTypes[measurement.ProductType]
	if !okProdType {
		return nil, fmt.Errorf("исполнение %q не определено", measurement.ProductType)
	}

	var result []SampleCalc

	for _, smp := range measurement.Samples {
		m, nMs := prodType.MeasurementByName(smp.Name)
		if nMs == -1 {
			return nil, fmt.Errorf("измерение %q не задано для исполнения %q", smp.Name, measurement.ProductType)
		}
		r := SampleCalc{Sample: smp, Calc: make([]Float64Ok, 16)}
		for place := range r.Calc {
			d := measureData{
				U:            smp.Productions[place].Value,
				I:            smp.Current,
				Q:            smp.Consumption,
				T:            smp.Temperature,
				prodType:     prodType,
				prodTypeName: measurement.ProductType,
				calc:         result,
				place:        place,
				pgs:          measurement.Pgs,
			}
			r.Calc[place].V, r.Calc[place].Ok = m.Calc(&d)
			if d.err != nil {
				return nil, fmt.Errorf("место %d: расчёт %s: %w", place, smp.Name, d.err)
			}
		}

		result = append(result, r)
	}
	return result, nil
}

func initProductTypes() {

	filename := filepath.Join(filepath.Dir(os.Args[0]), "lua", "sensel.lua")

	newMeasurement := func(name string, gas int, duration string, Calc CalcSamplesFunc) Measurement {
		dur, err := time.ParseDuration(duration)
		must.PanicIf(merry.Prepend(err, "duration"))
		return Measurement{
			Name:     name,
			Gas:      gas,
			Duration: dur,
			Calc:     Calc,
		}
	}

	LCalcSamples.SetGlobal("product", luar.New(LCalcSamples, func(name string, measurements ...Measurement) {
		if _, f := ProductTypes[name]; f {
			LCalcSamples.RaiseError("дублирование исполнения %q", name)
		}
		ProductTypes[name] = ProductType{
			Measurements: measurements,
		}
	}))
	LCalcSamples.SetGlobal("measure", luar.New(LCalcSamples, newMeasurement))
	must.PanicIf(merry.Prepend(LCalcSamples.DoFile(filename), filepath.Base(filename)))

	LCalcSamples.SetGlobal("product", lua.LNil)
	LCalcSamples.SetGlobal("measure", lua.LNil)

	testRandomSamples()
}

func testRandomSamples() {
	for prodTypeName, prodType := range ProductTypes {
		samples := make([]data.Sample, len(prodType.Measurements)-1)
		data.RandSamples(samples)

		for i := range samples {
			samples[i].Name = prodType.Measurements[i].Name
		}
		_, err := CalcSamples(data.Measurement{
			ProductType: prodTypeName,
			Pgs:         []float64{1, 2, 3, 4},
			Samples:     samples,
		})
		must.PanicIf(err)
	}
}

var (
	ProductTypes   = make(map[string]ProductType)
	LCalcSamples   = lua.NewState()
	nanMeasureData = prevMeasureData{
		Value: math.NaN(),
		U:     math.NaN(),
		I:     math.NaN(),
		Q:     math.NaN(),
		T:     math.NaN(),
	}
)
