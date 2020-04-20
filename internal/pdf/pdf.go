package pdf

import (
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/sensel/internal/calc"
	"github.com/fpawel/sensel/internal/data"
	"github.com/jung-kurt/gofpdf"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
)

func New(m data.Measurement, cs []calc.Column, cfg TableConfig, includeSamples bool) error {
	dir, err := prepareDir()
	if err != nil {
		return err
	}

	d := newDoc(m, cs, cfg, includeSamples)
	if err := saveAndShowDoc(d, dir, fmt.Sprintf("measure_%d", m.MeasurementID)); err != nil {
		return err
	}
	return nil
}

func newDoc(m data.Measurement, cs []calc.Column, tableConfig TableConfig, includeSamples bool) *gofpdf.Fpdf {
	//d := gofpdf.New("L", "mm", "A5", fontDir)
	doc := gofpdf.NewCustom(&gofpdf.InitType{
		OrientationStr: "L",
		UnitStr:        "mm",
		SizeStr:        "A5",
		Size:           gofpdf.SizeType{Wd: 210, Ht: 150},
		FontDirStr:     fontDir,
	})

	doc.AddFont("RobotoCondensed", "", "RobotoCondensed-Regular.json")
	doc.AddFont("RobotoCondensed", "B", "RobotoCondensed-Bold.json")
	doc.AddFont("RobotoCondensed", "I", "RobotoCondensed-Italic.json")
	doc.AddFont("RobotoCondensed", "BI", "RobotoCondensed-BoldItalic.json")
	doc.AddFont("RobotoCondensed-Light", "", "RobotoCondensed-Light.json")
	doc.AddFont("RobotoCondensed-Light", "I", "RobotoCondensed-LightItalic.json")
	doc.UnicodeTranslatorFromDescriptor("cp1251")
	doc.SetLineWidth(.1)
	doc.SetFillColor(225, 225, 225)
	doc.SetDrawColor(169, 169, 169)

	const (
		horizMargin = 3.
		topMargin   = 2.
	)

	doc.SetMargins(horizMargin, topMargin, horizMargin)
	doc.AddPage()

	pageWidth, _ := doc.GetPageSize()

	setFont := func(styleStr string, size float64) {
		doc.SetFont("RobotoCondensed", styleStr, size)
	}

	doc.SetX(horizMargin)

	setFont("B", 7)

	tr := doc.UnicodeTranslatorFromDescriptor("cp1251")
	doc.CellFormat(pageWidth-2.*horizMargin, 4, tr(fmt.Sprintf(
		"ЧЭ %s %s. Обмер № %d. Дата %s", m.Device, m.Kind, m.MeasurementID,
		m.CreatedAt.Format("02.01.06"))),
		"", 1, "C", false, 0, "")
	setFont("", tableConfig.FontSize)
	tbl := table{
		cfg: tableConfig,
		doc: doc,
	}
	tbl.cfg = tableConfig

	if includeSamples {
		tbl.rows = newTableRows1(m, cs)
		tbl.horizOffset = (pageWidth - tbl.width()) / 2.
		tbl.render()
	} else {
		tbl.rows = newTableRows2(m, cs)
		tbl.horizOffset = (pageWidth - tbl.width()) / 2.
		tbl.render()

		y := doc.GetY()

		tbl.rows = newTableMeasureSamples(m, true)
		doc.SetY(y + 1.)
		tbl.render()

		tbl.horizOffset += tbl.width() + 1

		doc.SetY(y + 1.)
		tbl.rows = newTableMeasureSamples(m, false)
		tbl.render()
	}
	return doc
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

var (
	fontDir = func() string {
		return filepath.Join(filepath.Dir(os.Args[0]), "assets", "fonts")
	}()
)
