package app

import (
	. "github.com/lxn/walk/declarative"
)

func (x *MeasurementViewModel) Columns() (xs []TableViewColumn) {
	appendResult := func(c TableViewColumn) {
		xs = append(xs, c)
	}
	appendResult(TableViewColumn{
		Name:  "Датчик",
		Title: "Датчик",
	})
	for _, smp := range x.M.Samples {
		appendResult(TableViewColumn{
			Name:  smp.Name,
			Title: smp.Name,
		})
	}
	return
}

const columnCurrentTitle = "U"
