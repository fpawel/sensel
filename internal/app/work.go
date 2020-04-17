package app

import (
	"context"
	"errors"
	"fmt"
	"github.com/ansel1/merry"
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

		// подключить все места
		if err := setupPlaceConnection(log, ctx, 0xFFFF); err != nil {
			return err
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

			if nSample > 0 {
				dataSmp.Br = measurement.Samples[nSample-1].Br
			}

			measurement.Samples = append(measurement.Samples, dataSmp)
			setMeasurementViewUISafe(measurement)

			// установить рабочее напряжение
			if err := setupTensionBar(log, ctx, smp.Tension); err != nil {
				return err
			}

			if err := setupCurrentBar(log, ctx, smp.Current); err != nil {
				return err
			}

			// установить газ
			if err := switchGas(log, ctx, smp.Gas); err != nil {
				//return err
			}

			ctxDelay, _ := context.WithTimeout(ctx, smp.Duration)
			if err := delay(log, ctxDelay, measurement, scheme); err != nil {
				return err
			}
			pause(ctx.Done(), cfg.Get().Voltmeter.PauseScan)

			if err := readAndSaveCurrentSample(log, ctx, &measurement); err != nil {
				return err
			}
		}
		return nil
	})
}

func readAndSaveCurrentSample(log comm.Logger, ctx context.Context, m *data.Measurement) error {
	dataSmp := &m.Samples[len(m.Samples)-1]
	err := readSample(log, ctx, dataSmp)
	if err != nil {
		return err
	}
	if err := data.SaveMeasurement(db, m); err != nil {
		return err
	}
	setMeasurementViewUISafe(*m)
	return nil
}

func delay(log comm.Logger, ctx context.Context, m data.Measurement, scheme []calc.Sample) error {

	smpIndex := len(m.Samples) - 1

	progress := func(t time.Time, d time.Duration) int {
		return int(100 * float64(time.Since(t)) / float64(d))
	}

	smp := scheme[smpIndex]
	tickUpdGui := time.NewTicker(time.Second * 2)
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
		default:
			err := readSample(log, ctx, &m.Samples[smpIndex])
			if err != nil {
				if merry.Is(err, context.DeadlineExceeded) {
					return nil
				}
				return err
			}
			setMeasurementViewUISafe(m)
		}
	}
}

func readSample(log comm.Logger, ctx context.Context, smp *data.Sample) error {
	if smp.BreakAll() {
		return errors.New("все элементы в планке оборваны")
	}
	if err := readVoltmeter(log, ctx, smp); err != nil {
		return err
	}

	// нет обрыва
	if smp.I > 0.006 {
		return nil
	}

	// установить напряжение 10В
	if err := setupTensionBar(log, ctx, 10); err != nil {
		return err
	}

	// найти обрыв
	if err := readBreak(log, ctx, smp); err != nil {
		return err
	}

	// установить рабочее напряжение
	if err := setupTensionBar(log, ctx, smp.Ub); err != nil {
		return err
	}
	smp.Tm = time.Now()
	return nil
}

func runReadVoltmeter() {
	runWork(func(ctx context.Context) error {
		for {
			var smp data.Sample
			err := readVoltmeter(log, ctx, &smp)
			if err != nil {
				return err
			}
			setSampleViewUISafe(smp)
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
		setSampleViewUISafe(smp)
		return nil
	})
}

func errNeedBytesCount(what string, n, len int) error {
	return fmt.Errorf("%s: ожидалось %d байт ответа, получено %d", what, n, len)
}
