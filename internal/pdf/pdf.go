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

func New(m data.Measurement, cs []calc.Column) error {
	dir, err := prepareDir()
	if err != nil {
		return err
	}
	d := gofpdf.New("L", "mm", "A5", fontDir)
	doDoc(d, m, cs)
	if err := saveAndShowDoc(d, dir, fmt.Sprintf("measure_%d", m.MeasurementID)); err != nil {
		return err
	}
	return nil
}

func doDoc(d *gofpdf.Fpdf, m data.Measurement, cs []calc.Column) {
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
		colWidth1     = 12.
		rowHeight     = 4.
		horizMargin   = 5.
		tableFontSize = 8
	)

	cellFormatGreyBackground := func(w, h float64, txtStr, borderStr string, ln int, alignStr string) {
		d.SetFillColor(230, 230, 230)
		d.CellFormat(w, h, txtStr, borderStr, ln, alignStr, true, 0, "")
		d.SetFillColor(255, 255, 255)
	}

	d.SetMargins(horizMargin, horizMargin, horizMargin)

	tr := d.UnicodeTranslatorFromDescriptor("cp1251")
	d.AddPageFormat("L", gofpdf.SizeType{1, 1})

	pageWidth, _ := d.GetPageSize()
	tableWidth := pageWidth - 2.*horizMargin

	setFont := func(styleStr string, size float64) {
		d.SetFont("RobotoCondensed", styleStr, size)
	}

	d.SetX(horizMargin)

	setFont("B", 9)

	d.CellFormat(tableWidth, 6, tr(fmt.Sprintf(
		"ЧЭ %s %s. Обмер № %d. Дата %s", m.Device, m.Kind, m.MeasurementID,
		m.CreatedAt.Format("02.01.06"))),
		"", 1, "L", false, 0, "")
	setFont("", tableFontSize)

	d.CellFormat(colWidth1, rowHeight, tr("Датчик"), "LRTB", 0, "C", false, 0, "")

	colWidth := (pageWidth - colWidth1 - 2.*horizMargin) / float64(len(cs))

	for _, c := range cs {
		d.CellFormat(colWidth, rowHeight, tr(c.Name), "LRTB", 0, "C", false, 0, "")
	}
	d.Ln(-1)

	for i := 0; i < 16; i++ {

		s := fmt.Sprintf("%d", i+1)

		if m.Br(i) {
			cellFormatGreyBackground(colWidth1, rowHeight, s, "LRTB", 0, "C")
		} else {
			d.CellFormat(colWidth1, rowHeight, s, "LRTB", 0, "C", false, 0, "")
		}

		for _, c := range cs {
			s := strconv.FormatFloat(c.Values[i].Value, 'f', c.Precision, 64)
			if c.IsErr(i) {
				cellFormatGreyBackground(colWidth, rowHeight, s, "LRTB", 0, "R")
			} else {
				d.CellFormat(colWidth, rowHeight, s, "LRTB", 0, "R", false, 0, "")
			}
		}
		d.Ln(-1)
	}
}

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
