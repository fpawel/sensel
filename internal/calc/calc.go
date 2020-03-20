package calc

import (
	"fmt"
	"github.com/fpawel/sensel/internal/data"
	lua "github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"
	"math"
	"sort"
	"time"
)

type C struct {
	l *lua.LState
	d mapProdTypes
}

type mapProdTypes map[string]map[string]prodType

type Sample struct {
	Gas      int
	Tension  float64
	Current  float64
	Duration time.Duration
}

type nameValueOk struct {
	Name      string
	Value     float64
	Ok        bool
	Precision int
}

type ValueOk struct {
	Value float64
	Ok    bool
}

func New(filename string) (C, error) {
	L := lua.NewState()

	var Import struct {
		Devices mapProdTypes
	}

	column := func(name string, value float64, ok bool, precision int) nameValueOk {
		return nameValueOk{
			Name:      name,
			Value:     value,
			Ok:        ok,
			Precision: precision,
		}
	}

	sample := func(gas int, strDuration string, U float64, I float64) Sample {
		dur, err := time.ParseDuration(strDuration)
		if err != nil {
			L.RaiseError("duration: %v", err)
		}
		return Sample{
			Gas:      gas,
			Tension:  U,
			Current:  I / 1000.,
			Duration: dur,
		}
	}

	L.SetGlobal("export", luar.New(L, &Import))
	L.SetGlobal("column", luar.New(L, column))
	L.SetGlobal("sample", luar.New(L, sample))

	if err := L.DoFile(filename); err != nil {
		return C{}, err
	}
	c := C{
		d: Import.Devices,
		l: L,
	}

	for device, m := range c.d {
		for kind, m := range m {
			_, err := m.Columns()
			if err != nil {
				return C{}, fmt.Errorf("%s: %s: %w", device, kind, err)
			}
		}
	}

	return c, nil
}

type Column struct {
	Values    [16]ValueOk
	Name      string
	Index     int
	Precision int
}

func (c C) ListDevices() (xs []string) {
	for s := range c.d {
		xs = append(xs, s)
	}
	return
}

func (c C) ListKinds(device string) (xs []string) {
	d, _ := c.d[device]
	for s := range d {
		xs = append(xs, s)
	}
	return
}

func (c C) CalculateMeasure(m data.Measurement) ([]Column, error) {
	mD, ok := c.d[m.Device]
	if !ok {
		return nil, fmt.Errorf("не найден тип прибора: %s %s", m.Device, m.Kind)
	}
	d, ok := mD[m.Kind]
	if !ok {
		return nil, fmt.Errorf("не найден тип прибора: %s %s", m.Device, m.Kind)
	}
	I := sliceLenNaN(len(d.Samples), m.I())
	T := sliceLenNaN(len(d.Samples), m.T())
	C := m.Pgs

	mCols := map[string]Column{}

	for i := 0; i < 16; i++ {
		U := sliceLenNaN(len(d.Samples), m.U(i))

		result, err := d.calculate(U, I, T, C)
		if err != nil {
			return nil, err
		}

		for columnIndex, c := range result {
			r, _ := mCols[c.Name]
			r.Name = c.Name
			r.Index = columnIndex
			r.Values[i].Value = c.Value
			r.Values[i].Ok = c.Ok
			r.Precision = c.Precision
			mCols[c.Name] = r
		}
	}

	var cols []Column
	for _, c := range mCols {
		cols = append(cols, c)
	}
	sort.Slice(cols, func(i, j int) bool {
		return cols[i].Index < cols[j].Index
	})

	return cols, nil
}

// sliceLenNaN - дополняет слайс а до длины n значениями NaN
func sliceLenNaN(n int, a []float64) []float64 {
	for len(a) < n {
		a = append(a, math.NaN())
	}
	return a
}
