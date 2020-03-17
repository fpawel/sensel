package calc

import "time"

type ProductType struct {
	Device   string
	Type     string
	Measures []Measure
}

type Measure struct {
	Gas      int
	Tension  float64
	Current  float64
	Duration time.Duration
}
