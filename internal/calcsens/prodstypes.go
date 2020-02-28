package calcsens

import (
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/sensel/internal/data"
	"github.com/fpawel/sensel/internal/pkg/must"
	lua "github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"
	"path/filepath"
	"time"
)

type ProductTypes struct {
	xs map[string]productType
	l  *lua.LState
}

type SampleCalc struct {
	data.Sample
	Calc []FloatOk
}

type FloatOk struct {
	Float float64
	Ok    bool
}

func NewProductTypes(filename string) (ProductTypes, error) {
	x := ProductTypes{
		xs: make(map[string]productType),
		l:  lua.NewState(),
	}

	newMeasurement := func(name string, gas int, duration string, Calc calcSamplesFunc) measurement {
		dur, err := time.ParseDuration(duration)
		must.PanicIf(merry.Prepend(err, "duration"))
		return measurement{
			Name:     name,
			Gas:      gas,
			Duration: dur,
			Calc:     Calc,
		}
	}

	x.l.SetGlobal("product", luar.New(x.l, func(name string, measurements ...measurement) {
		if _, f := x.xs[name]; f {
			x.l.RaiseError("дублирование исполнения %q", name)
		}
		x.xs[name] = productType{
			ms: measurements,
		}
	}))
	x.l.SetGlobal("measure", luar.New(x.l, newMeasurement))

	wrapErr := func(err error) (ProductTypes, error) {
		return ProductTypes{}, merry.Prepend(err, filepath.Base(filename))
	}

	if err := x.l.DoFile(filename); err != nil {
		return wrapErr(err)
	}

	x.l.SetGlobal("product", lua.LNil)
	x.l.SetGlobal("measure", lua.LNil)

	if err := x.testRandomSamples(); err != nil {
		return wrapErr(err)
	}

	return x, nil
}

func (x ProductTypes) CalcSamples(measurement data.Measurement) ([]SampleCalc, error) {
	prodType, okProdType := x.xs[measurement.ProductType]
	if !okProdType {
		return nil, fmt.Errorf("исполнение %q не определено", measurement.ProductType)
	}

	var calculated []SampleCalc

	for _, smp := range measurement.Samples {
		smp := smp
		m, foundMeasurement := prodType.measurementByName(smp.Name)
		if !foundMeasurement {
			return nil, fmt.Errorf("измерение %q не задано для исполнения %q", smp.Name, measurement.ProductType)
		}
		r := SampleCalc{Sample: smp, Calc: make([]FloatOk, 16)}
		for place := range r.Calc {
			d := measureData{
				U:           smp.Productions[place].Value,
				I:           smp.Current,
				Q:           smp.Consumption,
				T:           smp.Temperature,
				prodType:    prodType,
				calculated:  calculated,
				place:       place,
				dataMeasure: measurement,
			}
			r.Calc[place].Float, r.Calc[place].Ok = m.Calc(&d)
			if d.err != nil {
				return nil, fmt.Errorf("место %d: расчёт %s: %w", place, smp.Name, d.err)
			}
		}

		calculated = append(calculated, r)
	}
	return calculated, nil
}

func (x ProductTypes) ListProductTypes() (xs []string) {
	for name := range x.xs {
		xs = append(xs, name)
	}
	return
}

func (x ProductTypes) testRandomSamples() error {
	for prodTypeName, prodType := range x.xs {
		samples := make([]data.Sample, len(prodType.ms)-1)
		data.RandSamples(samples)

		for i := range samples {
			samples[i].Name = prodType.ms[i].Name
		}
		_, err := x.CalcSamples(data.Measurement{
			ProductType: prodTypeName,
			Pgs:         []float64{1, 2, 3, 4},
			Samples:     samples,
		})
		must.PanicIf(err)
	}
	return nil
}

type productType struct {
	ms []measurement
}

func (x productType) measurementByName(name string) (measurement, bool) {
	for _, x := range x.ms {
		if x.Name == name {
			return x, true
		}
	}
	return measurement{}, false
}
