package app

import (
	"fmt"
	"github.com/fpawel/sensel/internal/data"
	"github.com/fpawel/sensel/internal/pkg/must"
	"github.com/fpawel/sensel/internal/view/viewmeasure"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

type archiveWindow struct {
	win           *walk.MainWindow
	cb            *walk.ComboBox
	tv            *walk.TableView
	lblErr        *walk.LineEdit
	arch          []data.MeasurementInfo
	comboboxModel []string
	tvm           *viewmeasure.TableViewModel
}

func (x *archiveWindow) init() {
	must.PanicIf(data.ListArchive(db, &x.arch))
	var cbm []string
	for _, x := range x.arch {
		s := fmt.Sprintf("%10d   %s   %s %s   %q",
			x.MeasurementID, x.CreatedAt.Format("06.01.02 15:04"), x.Device, x.Kind, x.Name)
		cbm = append(cbm, s)
	}
	must.PanicIf(x.cb.SetModel(cbm))
	x.tvm = viewmeasure.NewMainTableViewModel(x.tv)
	//var m data.Measurement
	//_ = data.GetLastMeasurement(db, &m)
	//setMeasurement(m)
	_ = x.cb.SetCurrentIndex(0)
}

func (x *archiveWindow) setMeasurement(m data.Measurement) {
	calcCols, err := Calc.CalculateMeasure(m)
	if err != nil {
		must.PanicIf(x.lblErr.SetText(err.Error()))
		x.lblErr.SetVisible(true)
	} else {
		x.lblErr.SetVisible(false)
	}
	x.tvm.SetViewData(m, calcCols)
}

func (x *archiveWindow) Window() MainWindow {
	return MainWindow{
		Font: Font{
			Family:    "Segoe UI",
			PointSize: 12,
		},
		AssignTo: &x.win,
		Layout:   VBox{},
		Children: []Widget{
			ComboBox{
				AssignTo: &x.cb,
				Editable: false,
				OnCurrentIndexChanged: func() {
					if x.cb.CurrentIndex() < 0 || x.cb.CurrentIndex() >= len(x.arch) {
						return
					}
					var m data.Measurement
					m.MeasurementID = x.arch[x.cb.CurrentIndex()].MeasurementID
					must.PanicIf(data.GetMeasurement(db, &m))
					x.setMeasurement(m)
				},
			},
			TableView{
				AssignTo:                 &x.tv,
				ColumnsOrderable:         false,
				ColumnsSizable:           true,
				LastColumnStretched:      false,
				MultiSelection:           true,
				NotSortableByHeaderClick: true,
			},
			LineEdit{
				AssignTo: &x.lblErr,
			},
		},
	}
}
