package pdf

import (
	"github.com/jung-kurt/gofpdf"
)

type TableConfig struct {
	RowHeight      float64
	CellHorizSpace float64
	FontSize       float64
}

type tableCell struct {
	err bool
	text,
	border,
	align string
}

type tableCells []tableCell

type table struct {
	rows        []tableCells
	cfg         TableConfig
	doc         *gofpdf.Fpdf
	horizOffset float64
}

func (t table) render() {

	colWidths := t.colWidths()
	tr := t.doc.UnicodeTranslatorFromDescriptor("cp1251")
	for _, row := range t.rows {
		t.doc.SetX(t.horizOffset)
		for nCol, c := range row {
			text := tr(c.text)
			w := colWidths[nCol]
			ln := 0
			if nCol == len(colWidths)-1 {
				ln = 1
			}
			if c.err {
				t.doc.SetFillColor(230, 230, 230)
				t.doc.CellFormat(w, t.cfg.RowHeight, text, c.border, ln, c.align, true, 0, "")
				t.doc.SetFillColor(255, 255, 255)
				continue
			}
			t.doc.CellFormat(w, t.cfg.RowHeight, text, c.border, ln, c.align, false, 0, "")
		}
	}
}

func (t table) width() (w float64) {
	for i := 0; i < t.colCount(); i++ {
		w += t.col(i).width(t.doc, t.cfg.CellHorizSpace)
	}
	return
}

func (t table) colWidths() (xs []float64) {
	for i := 0; i < t.colCount(); i++ {
		xs = append(xs, t.col(i).width(t.doc, t.cfg.CellHorizSpace))
	}
	return
}

func (t table) colCount() int {
	if len(t.rows) == 0 {
		return 0
	}
	return len(t.rows[0])
}

func (t table) col(n int) (xs tableCells) {
	if n < 0 || n >= t.colCount() {
		panic("column out of range")
	}
	for _, row := range t.rows {
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
