package pdf

import (
	"database/sql"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/sensel/internal/calc"
	"github.com/fpawel/sensel/internal/data"
	"github.com/jung-kurt/gofpdf"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
)

func New(m data.Measurement, cs []calc.Column, cfg TableConfig) error {
	dir, err := prepareDir()
	if err != nil {
		return err
	}

	d := newDoc(m, cs, cfg)
	if err := saveAndShowDoc(d, dir, fmt.Sprintf("measure_%d", m.MeasurementID)); err != nil {
		return err
	}
	return nil
}

func newDoc(m data.Measurement, cs []calc.Column, tableConfig TableConfig) *gofpdf.Fpdf {
	//d := gofpdf.New("L", "mm", "A5", fontDir)
	d := gofpdf.NewCustom(&gofpdf.InitType{
		OrientationStr: "L",
		UnitStr:        "mm",
		SizeStr:        "A5",
		Size:           gofpdf.SizeType{Wd: 210, Ht: 150},
		FontDirStr:     fontDir,
	})

	d.AddFont("RobotoCondensed", "", "RobotoCondensed-Regular.json")
	d.AddFont("RobotoCondensed", "B", "RobotoCondensed-Bold.json")
	d.AddFont("RobotoCondensed", "I", "RobotoCondensed-Italic.json")
	d.AddFont("RobotoCondensed", "BI", "RobotoCondensed-BoldItalic.json")
	d.AddFont("RobotoCondensed-Light", "", "RobotoCondensed-Light.json")
	d.AddFont("RobotoCondensed-Light", "I", "RobotoCondensed-LightItalic.json")
	d.UnicodeTranslatorFromDescriptor("cp1251")
	d.SetLineWidth(.1)
	d.SetFillColor(225, 225, 225)
	d.SetDrawColor(169, 169, 169)

	const (
		horizMargin = 3.
		topMargin   = 2.
	)

	d.SetMargins(horizMargin, topMargin, horizMargin)
	d.AddPage()

	pageWidth, _ := d.GetPageSize()

	setFont := func(styleStr string, size float64) {
		d.SetFont("RobotoCondensed", styleStr, size)
	}

	d.SetX(horizMargin)

	setFont("B", 7)

	tr := d.UnicodeTranslatorFromDescriptor("cp1251")
	d.CellFormat(pageWidth-2.*horizMargin, 4, tr(fmt.Sprintf(
		"ЧЭ %s %s. Обмер № %d. Дата %s", m.Device, m.Kind, m.MeasurementID,
		m.CreatedAt.Format("02.01.06"))),
		"", 1, "C", false, 0, "")
	setFont("", tableConfig.FontSizePixels)
	tbl := newTable(m, cs)
	tbl.render(d, tableConfig)
	return d
}

func newTable(m data.Measurement, cs []calc.Column) (t table) {
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
	t = append(t, row)

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
				text:   strconv.FormatFloat(smp.U[i], 'f', 3, 64),
				border: "LRTB",
				align:  "R",
			})
		}

		for _, c := range cs {
			row = append(row, tableCell{
				err:    c.IsErr(i),
				text:   strconv.FormatFloat(c.Values[i].Value, 'f', c.Precision, 64),
				border: "LRTB",
				align:  "R",
			})
		}

		t = append(t, row)
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
				text:   strconv.FormatFloat(v, 'f', prec, 64),
				border: "LRTB",
				align:  "R",
			})
		}

		for range cs {
			row = append(row, tableCell{})
		}

		t = append(t, row)
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

//func newTableMeasure(m data.Measurement) (t table) {
//	row := []tableCell{
//		{
//			ok:     true,
//			text:   "Датчик",
//			border: "LRTB",
//			align:  "C",
//		},
//	}
//	for n := range m.Samples{
//		row = append(row, tableCell{
//			ok:     true,
//			text:   fmt.Sprintf("%d", n+1),
//			border: "LRTB",
//			align:  "C",
//		})
//	}
//
//	t = append(t, row)
//
//	for i := 0; i < 16; i++ {
//
//		row = []tableCell{
//			{
//				ok:     !m.Br(i),
//				text:   fmt.Sprintf("%d", i+1),
//				border: "LRTB",
//				align:  "C",
//			},
//		}
//		for _,smp := range m.Samples{
//			row = append(row, tableCell{
//				ok:     !smp.Br[i],
//				text:   strconv.FormatFloat(smp.U[i], 'f', 3, 64),
//				border: "LRTB",
//				align:  "R",
//			})
//		}
//
//		t = append(t, row)
//	}
//	return
//}

func prepareDir() (string, error) {

	dir := filepath.Join(filepath.Dir(os.Args[0]), "pdf")
	_ = os.RemoveAll(dir)
	_ = os.MkdirAll(dir, os.ModePerm)

	dir, err := ioutil.TempDir(dir, "~")
	if err != nil {
		return "", merry.WithMessage(err, "unable to create directory for pdf")
	}
	return dir, nil
}

func saveAndShowDoc(d *gofpdf.Fpdf, dir, fileName string) error {

	pdfFileName := filepath.Join(dir, fileName+".pdf")

	if err := d.OutputFileAndClose(pdfFileName); err != nil {
		return err
	}
	if err := exec.Command("explorer.exe", pdfFileName).Start(); err != nil {
		return err
	}
	return nil
}

func formatNullInt64(v sql.NullInt64) string {
	if v.Valid {
		return strconv.FormatInt(v.Int64, 10)
	}
	return ""
}

func formatNullFloat64(v sql.NullFloat64) string {
	if v.Valid {
		return formatFloat(v.Float64)
	}
	return ""
}

func formatFloat(v float64) string {
	return fmt.Sprintf("%v", v)
}

func formatNullFloat64Prec(v sql.NullFloat64, prec int) string {
	if v.Valid {
		return strconv.FormatFloat(v.Float64, 'f', prec, 64)
	}
	return ""
}

var (
	fontDir = func() string {
		return filepath.Join(filepath.Dir(os.Args[0]), "assets", "fonts")
	}()
)
