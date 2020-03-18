package calc

import "fmt"

type prodType struct {
	Samples   []Sample
	Calculate func(U, I, T, C []float64) []nameValueOk
}

func (x prodType) calculate(U, I, T, C []float64) (xs []nameValueOk, err error) {
	defer func() {
		x := recover()
		if x != nil {
			err = fmt.Errorf("%v", x)
		}
	}()
	xs = x.Calculate(U, I, T, C)
	return
}

func (x prodType) Columns() ([]string, error) {
	var C []float64
	for _, smp := range x.Samples {
		if len(C) <= smp.Gas {
			C = make([]float64, smp.Gas+1)
		}
	}
	dt := func() []float64 {
		return make([]float64, len(x.Samples))
	}
	I, U, T := dt(), dt(), dt()

	result, err := x.calculate(U, I, T, C)
	if err != nil {
		return nil, err
	}

	var xs []string
	for _, r := range result {
		xs = append(xs, r.Name)
	}
	return xs, nil
}
