package calcsens

import "time"

type ProductType struct {
	Name    string
	Columns []Column
}

type Column struct {
	Name     string
	Gas      int
	Duration time.Duration
}
