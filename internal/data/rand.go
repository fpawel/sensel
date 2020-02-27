package data

import (
	"math"
	"math/rand"
	"time"
)

func RandSamples(xs []Sample) {
	for i := range xs {
		xs[i].Productions = randProductions()
		xs[i].Temperature = rand3()
		xs[i].Current = rand3()
		xs[i].Consumption = rand3()
	}
}

func randProductions() []Production {
	xs := make([]Production, 16)
	for i := range xs {
		xs[i].Place = i
		xs[i].Break = rand.Float64() < 0.1
		xs[i].Value = rand3()
	}
	return xs
}

func rand3() float64 {
	return math.Round(rand.Float64()*1000) / 1000
}

func init() {
	rand.NewSource(time.Now().UnixNano())
}
