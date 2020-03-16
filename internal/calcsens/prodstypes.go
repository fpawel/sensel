package calcsens

import (
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/sensel/internal/data"
	"github.com/fpawel/sensel/internal/pkg/must"
	lua "github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"
	"math"
	"path/filepath"
	"sort"
	"time"
)

type C struct {
	xs map[string]productType
	l  *lua.LState
}

type ColumnCalculated struct {
	Name   string
	Values []FloatOk
}

type FloatOk struct {
	Float float64
	Ok    bool
}

func NewProductTypes(filename string) (C, error) {
	x := C{
		xs: make(map[string]productType),
		l:  lua.NewState(),
	}

	newMeasurement := func(name string, gas int, duration string, Calc calcSamplesFunc) column {
		dur, err := time.ParseDuration(duration)
		must.PanicIf(merry.Prepend(err, "duration"))
		return column{
			Name:     name,
			Gas:      gas,
			Duration: dur,
			Calc:     Calc,
		}
	}

	newColumn := func(name string, Calc calcSamplesFunc) column {
		return column{
			Name: name,
			Gas:  -1,
			Calc: Calc,
		}
	}

	x.l.SetGlobal("product", luar.New(x.l, func(name string, measurements ...column) {
		if _, f := x.xs[name]; f {
			x.l.RaiseError("дублирование исполнения %q", name)
		}
		x.xs[name] = productType{
			ms: measurements,
		}
	}))
	x.l.SetGlobal("measure", luar.New(x.l, newMeasurement))
	x.l.SetGlobal("column", luar.New(x.l, newColumn))

	wrapErr := func(err error) (C, error) {
		return C{}, merry.Prepend(err, filepath.Base(filename))
	}

	if err := x.l.DoFile(filename); err != nil {
		return wrapErr(err)
	}

	x.l.SetGlobal("product", lua.LNil)
	x.l.SetGlobal("measure", lua.LNil)
	x.l.SetGlobal("column", lua.LNil)

	if err := x.testRandomSamples(); err != nil {
		return wrapErr(err)
	}

	return x, nil
}

func (x C) ListProductTypes() (xs []string) {
	for name := range x.xs {
		xs = append(xs, name)
	}
	sort.Strings(xs)
	return
}

func (x C) GetFirstProductType() ProductType {
	for name, x := range x.xs {
		return x.ProductType(name)
	}
	panic("нет ни одного исполнения")
}

func (x C) GetProductTypeByName(name string) (ProductType, bool) {
	t, f := x.xs[name]
	if !f {
		return ProductType{}, false
	}
	return t.ProductType(name), true
}

func (x C) CalcSamples(measurement data.Measurement) ([]ColumnCalculated, ProductType, error) {

	prodType, okProdType := x.xs[measurement.ProductType]
	if !okProdType {
		return nil, ProductType{}, fmt.Errorf("исполнение %q не определено", measurement.ProductType)
	}

	calculated := make([]ColumnCalculated, len(prodType.ms))

	retProdType := prodType.ProductType(measurement.ProductType)

	for i, m := range prodType.ms {
		calculated[i].Values = make([]FloatOk, 16)

		for place := range calculated[i].Values {
			d := measureData{
				U:          math.NaN(),
				I:          math.NaN(),
				Q:          math.NaN(),
				T:          math.NaN(),
				prodType:   prodType,
				calculated: calculated,
				place:      place,
				d:          measurement,
			}
			if smp, smpFound := measurement.Samples.GetSampleByName(m.Name); smpFound {
				d.U = smp.Productions[place].Value
				d.I = smp.CurrentBar
				d.Q = smp.Consumption
				d.T = smp.Temperature
			}
			v, ok := m.Calc(&d)
			calculated[i].Values[place] = FloatOk{
				Float: v,
				Ok:    ok,
			}
			if d.err != nil {
				return nil, retProdType, fmt.Errorf("место %d: расчёт %s: %w", place, m.Name, d.err)
			}
		}
	}
	return calculated, retProdType, nil
}

func (x C) testRandomSamples() error {
	for prodTypeName, prodType := range x.xs {
		samples := make([]data.Sample, len(prodType.ms)-1)
		data.RandSamples(samples)

		for i := range samples {
			samples[i].Name = prodType.ms[i].Name
		}
		_, _, err := x.CalcSamples(data.Measurement{
			MeasurementInfo: data.MeasurementInfo{
				ProductType: prodTypeName,
			},
			MeasurementData: data.MeasurementData{
				Pgs:     []float64{1, 2, 3, 4},
				Samples: samples,
			},
		})
		return err
	}
	return nil
}

type productType struct {
	ms []column
}

func (x productType) measurementByName(name string) (column, bool) {
	for _, x := range x.ms {
		if x.Name == name {
			return x, true
		}
	}
	return column{}, false
}

func (x productType) ProductType(name string) ProductType {
	r := ProductType{Name: name}
	for _, x := range x.ms {
		r.Columns = append(r.Columns, x.Column())
	}
	return r
}
