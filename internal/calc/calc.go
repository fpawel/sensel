package calc

import (
	lua "github.com/yuin/gopher-lua"
	"time"
)

type C struct {
	l *lua.LState
	d map[string]map[string]ProductType
}

type ProductType struct {
	Samples   []Sample
	Calculate func(U, I, T, C []float64) []NameValueOk
}

type Sample struct {
	Gas      int
	Tension  float64
	Current  float64
	Duration time.Duration
}

type NameValueOk struct {
	Name  string
	Value float64
	Ok    bool
}
