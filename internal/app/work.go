package app

import (
	"context"
	"fmt"
	"github.com/fpawel/comm"
	"github.com/fpawel/sensel/internal/calc"
	"github.com/fpawel/sensel/internal/cfg"
	"github.com/fpawel/sensel/internal/data"
	"time"
)

func runMeasure(measurement data.Measurement) {

	runWork(func(ctx context.Context) error {

		defer func() {
			setStatusOkSync(labelWorkStatus, "Отключение газа по окончании обмера")
			if err := switchOffGas(log, context.Background()); err != nil {
				log.PrintErr(err)
			}
		}()

		scheme, err := Calc.GetProductTypeMeasurementScheme(measurement.Device, measurement.Kind)
		if err != nil {
			return fmt.Errorf("%s: %s: %w", measurement.Device, measurement.Kind, err)
		}

		for nSample, smp := range scheme {
			if ctx.Err() != nil {
				return ctx.Err()
			}

			setStatusOkSync(labelWorkStatus, fmt.Sprintf("Измерение %d газ=%d U=%g I=%g",
				nSample+1, smp.Gas, smp.Tension, smp.Current))

			dataSmp := data.Sample{
				Gas: smp.Gas,
				Ub:  smp.Tension,
			}
			if err := setupMeasurement(log, ctx, smp); err != nil {
				return err
			}
			if err := readBar(log, ctx, &dataSmp); err != nil {
				return err
			}
			measurement.Samples = append(measurement.Samples, dataSmp)
			setMeasurement(measurement)

			ctxDelay, _ := context.WithTimeout(ctx, smp.Duration)
			if err := delay(log, ctxDelay, &measurement, scheme); err != nil {
				return err
			}

			if err := readAndSaveCurrentSample(log, ctx, &measurement); err != nil {
				return err
			}
		}
		return nil
	})
}

func readAndSaveCurrentSample(log comm.Logger, ctx context.Context, m *data.Measurement) error {
	dataSmp := &m.Samples[len(m.Samples)-1]
	err := readBar(log, ctx, dataSmp)
	if err != nil {
		return err
	}
	if err := data.SaveMeasurement(db, m); err != nil {
		return err
	}
	setMeasurement(*m)
	return nil
}

func delay(log comm.Logger, ctx context.Context, m *data.Measurement, scheme []calc.Sample) error {

	smpIndex := len(m.Samples) - 1

	progress := func(t time.Time, d time.Duration) int {
		return int(100 * float64(time.Since(t)) / float64(d))
	}

	smp := scheme[smpIndex]
	tickUpdGui := time.NewTicker(time.Second * 2)
	tickReadSample := time.NewTicker(cfg.Get().ReadSampleInterval)
	startTime := time.Now()

	upd := func() {
		d1 := progress(startTime, smp.Duration)
		totalDur := calc.GetTotalMeasureDuration(scheme)
		d2 := progress(m.CreatedAt, totalDur)
		progressBarCurrentWork.SetValue(d1)
		progressBarTotalWork.SetValue(d2)
		s := fmt.Sprintf("%s из %s %d%s", time.Since(startTime), smp.Duration, d1, "%")
		_ = labelCurrentDelay.SetText(s)
		s = fmt.Sprintf("Общий прогресс выполнения %s из %s %d%s",
			time.Since(m.CreatedAt), totalDur, d2, "%")
		_ = labelTotalDelay.SetText(s)
	}

	appWindow.Synchronize(func() {
		progressBarCurrentWork.SetVisible(true)
		progressBarTotalWork.SetVisible(true)

		progressBarCurrentWork.SetRange(0, 100)
		progressBarTotalWork.SetRange(0, 100)
		progressBarTotalWork.SetRange(0, 100)
		labelCurrentDelay.SetVisible(true)
		labelTotalDelay.SetVisible(true)
		upd()
	})

	defer func() {
		tickUpdGui.Stop()
		appWindow.Synchronize(func() {
			progressBarCurrentWork.SetVisible(false)
			progressBarTotalWork.SetVisible(false)
			labelCurrentDelay.SetVisible(false)
			labelTotalDelay.SetVisible(false)
		})
	}()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-tickUpdGui.C:
			appWindow.Synchronize(upd)
		case <-tickReadSample.C:
			if err := readAndSaveCurrentSample(log, ctx, m); err != nil {
				return err
			}
		}
	}
}

func setupMeasurement(log comm.Logger, ctx context.Context, sampleScheme calc.Sample) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if err := setupCurrentBar(log, ctx, sampleScheme.Current); err != nil {
		return err
	}
	if err := setupTensionBar(log, ctx, sampleScheme.Tension); err != nil {
		return err
	}
	if err := switchGas(log, ctx, sampleScheme.Gas); err != nil {
		//return err
	}
	return nil
}

func runReadSample() {
	runWork(func(ctx context.Context) error {
		for {
			var smp data.Sample
			err := readBar(log, ctx, &smp)
			if err != nil {
				return err
			}
			appWindow.Synchronize(func() {
				labelCalcErr.SetVisible(false)
				getMeasureTableViewModel().SetViewData(data.Measurement{
					MeasurementData: data.MeasurementData{
						Samples: []data.Sample{smp},
					},
				}, nil)
			})
		}
	})
}

func runSearchBreak() {
	runWork(func(ctx context.Context) error {
		// установить напряжение питания 10 В
		if err := setupTensionBar(log, ctx, 10); err != nil {
			return err
		}
		var smp data.Sample
		err := readBreak(log, ctx, &smp)
		if err != nil {
			return err
		}
		appWindow.Synchronize(func() {
			labelCalcErr.SetVisible(false)
			getMeasureTableViewModel().SetViewData(data.Measurement{
				MeasurementData: data.MeasurementData{
					Samples: []data.Sample{smp},
				},
			}, nil)
		})
		return nil
	})
}

func readBar(log comm.Logger, ctx context.Context, smp *data.Sample) error {
	//if err := readBreak(log, ctx, smp); err != nil {
	//	return err
	//}
	if err := readVoltmeter(log, ctx, smp); err != nil {
		return err
	}
	smp.Tm = time.Now()
	return nil
}

func errNeedBytesCount(what string, n, len int) error {
	return fmt.Errorf("%s: ожидалось %d байт ответа, получено %d", what, n, len)
}
