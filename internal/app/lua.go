package app

import (
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/sensel/internal/data"
	"github.com/fpawel/sensel/internal/pkg/must"
	lua "github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"
	"os"
	"path/filepath"
	"time"
)

type ProductType struct {
	Measurements []Measurement
}

func (x ProductType) MeasurementByName(name string) (Measurement, bool) {
	for _, x := range x.Measurements {
		if x.Name == name {
			return x, true
		}
	}
	return Measurement{}, false
}

type Measurement struct {
	Name     string
	Gas      int
	Duration time.Duration
	Calc     CalcSamplesFunc
}

type CalcSamplesFunc = func(M, V, I, Q, T MapStrFloat) (float64, bool)

type MapStrFloat map[string]float64

type SampleCalc struct {
	data.Sample
	Calc []Float64Ok
}

type Float64Ok struct {
	V  float64
	Ok bool
}

func CalcSamples(productTypeName string, samples []data.Sample) ([]SampleCalc, error) {
	prodType, okProdType := ProductTypes[productTypeName]
	if !okProdType {
		return nil, fmt.Errorf("исполнение %q не определено", productTypeName)
	}

	I := make(MapStrFloat)
	Q := make(MapStrFloat)
	T := make(MapStrFloat)
	result := make([]SampleCalc, len(samples))

	for i, smp := range samples {
		result[i].Sample = smp
		name := smp.Name
		I[name] = smp.Current
		Q[name] = smp.Consumption
		T[name] = smp.Temperature
		if _, f := prodType.MeasurementByName(name); !f {
			return nil, fmt.Errorf("измерение %q не задано для исполнения %q", name, productTypeName)
		}
	}
	for n := 0; n < 16; n++ {
		M := make(MapStrFloat)
		V := make(MapStrFloat)
		for i, smp := range result {
			m, _ := prodType.MeasurementByName(smp.Name)
			M[smp.Name] = smp.Productions[n].Value
			var v Float64Ok
			v.V, v.Ok = m.Calc(M, V, I, Q, T)
			V[smp.Name] = v.V
			result[i].Calc = append(result[i].Calc, v)
		}
	}
	return result, nil
}

func initProductTypes() {

	filename := filepath.Join(filepath.Dir(os.Args[0]), "lua", "sensel.lua")
	var prodTypeName string

	addMeasurement := func(name string, gas int, duration string, Calc CalcSamplesFunc) {
		dur, err := time.ParseDuration(duration)
		must.PanicIf(merry.Prepend(err, "duration"))
		m := Measurement{
			Name:     name,
			Gas:      gas,
			Duration: dur,
			Calc:     Calc,
		}
		prodType := ProductTypes[prodTypeName]
		prodType.Measurements = append(prodType.Measurements, m)
		ProductTypes[prodTypeName] = prodType
	}

	LCalcSamples.SetGlobal("Product", luar.New(LCalcSamples, func(name string) {
		if _, f := ProductTypes[name]; f {
			LCalcSamples.RaiseError("дублирование исполнения %q", name)
		}
		ProductTypes[name] = ProductType{}
		prodTypeName = name
	}))
	LCalcSamples.SetGlobal("Measurement", luar.New(LCalcSamples, addMeasurement))
	must.PanicIf(merry.Prepend(LCalcSamples.DoFile(filename), filepath.Base(filename)))

	LCalcSamples.SetGlobal("Product", lua.LNil)
	LCalcSamples.SetGlobal("Measurement", lua.LNil)

	testRandomSamples()
}

func testRandomSamples() {
	for prodTypeName, prodType := range ProductTypes {
		samples := make([]data.Sample, len(prodType.Measurements))
		data.RandSamples(samples)

		for i := range samples {
			samples[i].Name = prodType.Measurements[i].Name
		}
		_, err := CalcSamples(prodTypeName, samples)
		must.PanicIf(err)
	}
}

var (
	ProductTypes = make(map[string]ProductType)
	LCalcSamples = lua.NewState()
)
