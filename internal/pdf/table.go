package pdf

import (
	"github.com/jung-kurt/gofpdf"
)

type tableCell struct {
	err bool
	text,
	border,
	align string
}

type tableCells []tableCell

type table []tableCells

type TableConfig struct {
	RowHeightMM      float64
	CellHorizSpaceMM float64
	FontSizePixels   float64
}

func (t table) render(d *gofpdf.Fpdf, cfg TableConfig) {

	pageWidth, _ := d.GetPageSize()

	colWidths := t.colWidths(d, cfg.CellHorizSpaceMM)
	tr := d.UnicodeTranslatorFromDescriptor("cp1251")
	tableWidth := t.width(d, cfg.CellHorizSpaceMM)
	x0 := (pageWidth - tableWidth) / 2.
	for _, row := range t {
		d.SetX(x0)
		for nCol, c := range row {
			text := tr(c.text)
			w := colWidths[nCol]
			ln := 0
			if nCol == len(colWidths)-1 {
				ln = 1
			}
			if c.err {
				d.SetFillColor(230, 230, 230)
				d.CellFormat(w, cfg.RowHeightMM, text, c.border, ln, c.align, true, 0, "")
				d.SetFillColor(255, 255, 255)
				continue
			}
			d.CellFormat(w, cfg.RowHeightMM, text, c.border, ln, c.align, false, 0, "")
		}
	}
}

func (t table) width(d *gofpdf.Fpdf, cellHorizSpace float64) (w float64) {
	for i := 0; i < t.colCount(); i++ {
		w += t.col(i).width(d, cellHorizSpace)
	}
	return
}

func (t table) colWidths(d *gofpdf.Fpdf, cellHorizSpace float64) (xs []float64) {
	for i := 0; i < t.colCount(); i++ {
		xs = append(xs, t.col(i).width(d, cellHorizSpace))
	}
	return
}

func (t table) colCount() int {
	if len(t) == 0 {
		return 0
	}
	return len(t[0])
}

func (t table) col(n int) (xs tableCells) {
	if n < 0 || n >= t.colCount() {
		panic("column out of range")
	}
	for _, row := range t {
		xs = append(xs, row[n])
	}
	return
}

func (tc tableCells) width(d *gofpdf.Fpdf, cellHorizSpace float64) (w float64) {
	tr := d.UnicodeTranslatorFromDescriptor("cp1251")
	for _, c := range tc {
		x := d.GetStringWidth(tr(c.text)) + cellHorizSpace
		if x > w {
			w = x
		}
	}
	return
}
