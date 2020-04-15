package pdf

import (
	"fmt"
	"github.com/fpawel/sensel/internal/calc"
	"github.com/fpawel/sensel/internal/data"
	"github.com/fpawel/sensel/internal/pkg"
)

func newTableRows1(m data.Measurement, cs []calc.Column) (rows []tableCells) {
	row := []tableCell{
		{
			text:   "Датчик",
			border: "LRTB",
			align:  "C",
		},
	}
	for n := range m.Samples {
		row = append(row, tableCell{
			text:   fmt.Sprintf("U%d", n+1),
			border: "LRTB",
			align:  "C",
		})
	}

	for _, c := range cs {
		row = append(row, tableCell{
			text:   c.Name,
			border: "LRTB",
			align:  "C",
		})
	}
	rows = append(rows, row)

	for i := 0; i < 16; i++ {

		row = []tableCell{
			{
				err:    m.Br(i),
				text:   fmt.Sprintf("%d", i+1),
				border: "LRTB",
				align:  "C",
			},
		}
		for _, smp := range m.Samples {
			row = append(row, tableCell{
				err:    smp.Br[i],
				text:   pkg.FormatFloat(smp.U[i], 3),
				border: "LRTB",
				align:  "R",
			})
		}

		for _, c := range cs {
			row = append(row, tableCell{
				err:    c.IsErr(i),
				text:   pkg.FormatFloat(c.Values[i].Value, c.Precision),
				border: "LRTB",
				align:  "R",
			})
		}

		rows = append(rows, row)
	}

	newMeasureRow := func(title string, f func(smp data.Sample) (float64, int)) {
		row := []tableCell{
			{
				border: "LRTB",
				text:   title,
				align:  "C",
			},
		}

		for _, smp := range m.Samples {
			v, prec := f(smp)
			row = append(row, tableCell{
				text:   pkg.FormatFloat(v, prec),
				border: "LRTB",
				align:  "R",
			})
		}

		for range cs {
			row = append(row, tableCell{})
		}

		rows = append(rows, row)
	}

	newMeasureRow("U,мВ", func(smp data.Sample) (float64, int) {
		return smp.Ub, 2
	})
	newMeasureRow("I,мA", func(smp data.Sample) (float64, int) {
		return smp.I, 3
	})
	newMeasureRow("Т\"C", func(smp data.Sample) (float64, int) {
		return smp.T, 1
	})
	newMeasureRow("Q", func(smp data.Sample) (float64, int) {
		return smp.Q, 3
	})
	newMeasureRow("Газ", func(smp data.Sample) (float64, int) {
		return float64(smp.Gas + 1), 0
	})
	newMeasureRow("ПГС", func(smp data.Sample) (float64, int) {
		return m.Pgs[smp.Gas], 2
	})
	return
}

func newTableRows2(m data.Measurement, cs []calc.Column) (rows []tableCells) {
	row := []tableCell{
		{
			text:   "Датчик",
			border: "LRTB",
			align:  "C",
		},
	}

	for _, c := range cs {
		row = append(row, tableCell{
			text:   c.Name,
			border: "LRTB",
			align:  "C",
		})
	}
	rows = append(rows, row)

	for i := 0; i < 16; i++ {

		row = []tableCell{
			{
				err:    m.Br(i),
				text:   fmt.Sprintf("%d", i+1),
				border: "LRTB",
				align:  "C",
			},
		}

		for _, c := range cs {
			row = append(row, tableCell{
				err:    c.IsErr(i),
				text:   pkg.FormatFloat(c.Values[i].Value, c.Precision),
				border: "LRTB",
				align:  "R",
			})
		}

		rows = append(rows, row)
	}

	return
}

func newTableMeasureSamples(m data.Measurement) (rows []tableCells) {
	row := tableCells{
		{
			text:   "Замер",
			border: "LRTB",
			align:  "C",
		},
	}
	for n := range m.Samples {
		row = append(row, tableCell{
			text:   fmt.Sprintf("%d", n+1),
			border: "LRTB",
			align:  "C",
		})
	}
	rows = append(rows, row)

	newMeasureRow := func(title string, f func(smp data.Sample) (float64, int)) {
		row := []tableCell{
			{
				border: "LRTB",
				text:   title,
				align:  "C",
			},
		}

		for _, smp := range m.Samples {
			v, prec := f(smp)
			row = append(row, tableCell{
				text:   pkg.FormatFloat(v, prec),
				border: "LRTB",
				align:  "R",
			})
		}
		rows = append(rows, row)
	}

	newMeasureRow("U,мВ", func(smp data.Sample) (float64, int) {
		return smp.Ub, 2
	})
	newMeasureRow("I,мA", func(smp data.Sample) (float64, int) {
		return smp.I, 3
	})
	newMeasureRow("Т\"C", func(smp data.Sample) (float64, int) {
		return smp.T, 1
	})
	newMeasureRow("Q", func(smp data.Sample) (float64, int) {
		return smp.Q, 3
	})
	newMeasureRow("Газ", func(smp data.Sample) (float64, int) {
		return float64(smp.Gas + 1), 0
	})
	newMeasureRow("ПГС", func(smp data.Sample) (float64, int) {
		return m.Pgs[smp.Gas], 2
	})
	return
}
