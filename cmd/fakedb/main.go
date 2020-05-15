package main

import (
	"fmt"
	"github.com/fpawel/sensel/internal/calc"
	"github.com/fpawel/sensel/internal/data"
	"github.com/fpawel/sensel/internal/pkg/must"
	"github.com/schollz/progressbar/v3"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"time"
)

func main() {
	exeDir := filepath.Dir(os.Args[0])
	scheme, err := calc.New(filepath.Join(exeDir, "lua", "sensel.lua"))
	must.PanicIf(err)

	// соединение с базой данных
	dbFilename := filepath.Join(exeDir, "sensel.sqlite")
	db, err := data.Open(dbFilename)
	must.PanicIf(err)

	defer func() {
		must.PanicIf(db.Close())
	}()

	_, err = db.Exec(`DELETE FROM measurement WHERE TRUE`)
	must.PanicIf(err)

	tm := time.Now()

	n := 0
	for _, device := range scheme.ListDevices() {
		for range scheme.ListKinds(device) {
			n++
		}
	}

	bar := progressbar.NewOptions(100*n, progressbar.OptionSetPredictTime(true))

	for _, device := range scheme.ListDevices() {
		for _, kind := range scheme.ListKinds(device) {

			calcSamples, err := scheme.GetProductTypeMeasurementScheme(device, kind)
			must.PanicIf(err)

			for i := 0; i < 10; i++ {
				tm = tm.Add(-time.Hour * 24)
				for j := 0; j < 10; j++ {
					samples := make([]data.Sample, len(calcSamples))
					randSamples(samples)
					for i := range samples {
						samples[i].Tm = tm
						tm = tm.Add(-time.Hour)
					}
					m := data.Measurement{
						MeasurementInfo: data.MeasurementInfo{
							CreatedAt: tm,
							Device:    device,
							Kind:      kind,
							Name:      fmt.Sprintf("%v", tm),
						},
						MeasurementData: data.MeasurementData{
							Pgs:     []float64{rand3(), rand3(), rand3(), rand3()},
							Samples: samples,
						},
					}
					tm = tm.Add(-time.Hour)
					must.PanicIf(data.SaveMeasurement(db, &m))
					bar.Add(1)
				}
			}
		}
	}
}

func randSamples(xs []data.Sample) {
	for i := range xs {
		xs[i].Gas = i%4 + 1
		xs[i].T = rand3()
		xs[i].I = rand3()
		xs[i].Q = rand3()
		for j := range xs[i].U {
			xs[i].U[j] = rand3()
			xs[i].Br[j] = rand.Float64() < 0.3
		}
	}
}

func rand3() float64 {
	return math.Round(rand.Float64()*1000) / 1000
}

func init() {
	rand.NewSource(time.Now().UnixNano())
}
