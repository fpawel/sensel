package viewarch

import (
	"github.com/fpawel/sensel/internal/data"
	"github.com/lxn/walk"
)

type TableViewModel struct {
	walk.TableModelBase
	d []data.MeasurementInfo
}

var _ walk.TableModel = new(TableViewModel)

func (x *TableViewModel) ViewData() []data.MeasurementInfo {
	return x.d
}

func (x *TableViewModel) SetViewData(d []data.MeasurementInfo) {
	x.d = d
	x.PublishRowsReset()
}

func (x *TableViewModel) RowCount() int {
	return len(x.d)
}

func (x *TableViewModel) Value(row, col int) interface{} {
	if row < 0 || row >= len(x.d) {
		return ""
	}
	d := x.d[row]
	vd := []interface{}{
		d.MeasurementID,
		d.CreatedAt.Format("2006"),
		d.CreatedAt.Format("01"),
		d.CreatedAt.Format("06"),
		d.CreatedAt.Format("15:04"),
		d.Device + " " + d.Kind,
		d.Name,
	}
	if col < 0 || col >= len(vd) {
		return ""
	}
	return vd[col]
}
