package app

import (
	"context"
	"errors"
	"fmt"
	"github.com/fpawel/comm"
	"github.com/fpawel/sensel/internal/calc"
	"github.com/fpawel/sensel/internal/cfg"
	"github.com/fpawel/sensel/internal/data"
	"github.com/fpawel/sensel/internal/pkg"
	"github.com/lxn/walk"
	"os/exec"
	"time"
)

func runMeasure(measurement data.Measurement) {

	runWork(func(ctx context.Context) error {

		defer func() {
			setStatusOkSync(labelWorkStatus, "Подача воздуха по окончании обмера")
			if err := switchOffGas(log, context.Background()); err != nil {
				log.PrintErr(err)
			}

			setStatusOkSync(labelWorkStatus, "Установка тока 0 по окончании обмера")
			if err := setupCurrentBar(log, context.Background(), 0); err != nil {
				log.PrintErr(err)
			}

			setStatusOkSync(labelWorkStatus, "Установка напряжения 0 по окончании обмера")
			if err := setupTensionBar(log, context.Background(), 0); err != nil {
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

			measurement.Samples = append(measurement.Samples, data.Sample{
				Gas: smp.Gas,
				Ub:  smp.Tension,
				Ib:  smp.Current,
			})

			dataSmp := &measurement.Samples[nSample]

			if nSample > 0 {
				dataSmp.Br = measurement.Samples[nSample-1].Br
			}

			// установить рабочее напряжение
			if err := setupTensionBar(log, ctx, smp.Tension); err != nil {
				return err
			}

			// установить рабочий ток
			if err := setupCurrentBar(log, ctx, smp.Current); err != nil {
				return err
			}

			// установить газ
			if err := switchGas(log, ctx, smp.Gas); err != nil {
				return err
			}

			if err := readSample(log, ctx, dataSmp); err != nil {
				return err
			}

			setMeasurementViewUISafe(measurement)

			if dataSmp.I < 0.002 {
				// найти обрыв
				if err := processBreak(log, ctx, &measurement, dataSmp); err != nil {
					return err
				}
			}

			if err := delay(log, ctx, measurement, scheme); err != nil {
				return err
			}

			if err := readSample(log, ctx, dataSmp); err != nil {
				return err
			}
			if err := data.SaveMeasurement(db, &measurement); err != nil {
				return err
			}

			setMeasurementViewUISafe(measurement)
		}

		if err := printMeasurement(measurement); err != nil {
			return err
		}

		walk.MsgBox(appWindow, "Обмер завершён", fmt.Sprintf("Обмер %d завершён успешно.", measurement.MeasurementID), walk.MsgBoxIconInformation)

		return nil
	})
}

func printMeasurement(m data.Measurement) error {
	filename, err := newPdf(m)
	if err != nil {
		return err
	}

	if err := exec.Command("PDFtoPrinter", filename, cfg.Get().Printer).Start(); err != nil {
		return err
	}
	return nil
}

func delay(log comm.Logger, ctx context.Context, m data.Measurement, scheme []calc.Sample) error {
	smpIndex := len(m.Samples) - 1

	progress := func(t time.Time, d time.Duration) int {
		return int(100 * float64(time.Since(t)) / float64(d))
	}

	smp := scheme[smpIndex]

	startTime := time.Now()

	upd := func() {
		d1 := progress(startTime, smp.Duration)
		totalDur := getMeasureDuration(scheme)
		var elapsed time.Duration
		if smpIndex > 0 {
			elapsed = getMeasureDuration(scheme[:smpIndex])
		}
		time0 := startTime.Add(-elapsed)

		d2 := progress(time0, totalDur)

		progressBarCurrentWork.SetValue(d1)
		progressBarTotalWork.SetValue(d2)

		s := fmt.Sprintf("%s : %s - %s : %d%s",
			startTime.Format("15:04:05"),
			pkg.FormatDuration(time.Since(startTime)),
			pkg.FormatDuration(smp.Duration), d1, "%")
		_ = labelCurrentDelay.SetText(s)

		s = fmt.Sprintf("Обмер : %s - %s : %d%s",
			pkg.FormatDuration(time.Since(time0)),
			pkg.FormatDuration(totalDur), d2, "%")
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
		appWindow.Synchronize(func() {
			progressBarCurrentWork.SetVisible(false)
			progressBarTotalWork.SetVisible(false)
			labelCurrentDelay.SetVisible(false)
			labelTotalDelay.SetVisible(false)
		})
	}()

	ctxDelay, cancelDelay := context.WithTimeout(ctx, smp.Duration)
	defer cancelDelay()

	go func() {
		tickUpdGui := time.NewTicker(time.Second * 1)
		defer tickUpdGui.Stop()
		for {
			select {
			case <-ctxDelay.Done():
				return
			case <-tickUpdGui.C:
				appWindow.Synchronize(upd)
			}
		}
	}()

	conf := cfg.Get()
	tickVoltmeter := time.NewTicker(conf.Voltmeter.PauseMeasureScan)
	tickConsumption := time.NewTicker(time.Second)
	defer func() {
		tickVoltmeter.Stop()
		tickConsumption.Stop()
	}()

	for {

		if ctx.Err() != nil {
			return ctx.Err()
		}
		smp := &m.Samples[smpIndex]
		select {
		case <-ctxDelay.Done():
			return nil
		case <-tickVoltmeter.C:
			if err := readSample(log, ctx, smp); err != nil {
				return err
			}
			setMeasurementViewUISafe(m)
		case <-tickConsumption.C:
			var err error
			if smp.Q, err = readGasConsumption(log, ctx); err != nil {
				return err
			}
			setMeasurementViewUISafe(m)
		}
	}
}

func getMeasureDuration(xs []calc.Sample) (d time.Duration) {
	for _, x := range xs {
		d += x.Duration
	}
	return
}

func readSample(log comm.Logger, ctx context.Context, smp *data.Sample) error {
	if smp.BreakAll() {
		return errors.New("все элементы в планке оборваны")
	}

	var err error
	smp.Q, err = readGasConsumption(log, ctx)
	if err != nil {
		return err
	}

	if err := readVoltmeter(log, ctx, smp); err != nil {
		return err
	}

	smp.Tm = time.Now()

	return nil
}

func runCheckConnection() {
	runWork(func(ctx context.Context) error {
		var smp data.Sample
		// проверить вольтметр
		if err := readVoltmeter(log, ctx, &smp); err != nil {
			return err
		}
		// отключить все места
		if err := setupPlaceConnection(log, ctx, 0); err != nil {
			return err
		}
		// установить напряжение 10В
		if err := setupTensionBar(log, ctx, 10); err != nil {
			return err
		}
		// установить ток 0
		if err := setupCurrentBar(log, ctx, 0); err != nil {
			return err
		}
		// отключить газ
		if err := switchGas(log, ctx, 0); err != nil {
			return err
		}
		walk.MsgBox(appWindow, "Связь установлена", "Связь установлена. Оборудование отвечает.", walk.MsgBoxIconInformation)
		return nil
	})
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

func runReadSample() {
	runWork(func(ctx context.Context) error {
		for {
			var smp data.Sample
			err := readSample(log, ctx, &smp)
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
		err := processBreak(log, ctx, nil, &smp)
		if err != nil {
			return err
		}
		setSampleViewUISafe(smp)
		return nil
	})
}

func runReadConsumption() {
	tableViewMeasure.SetVisible(false)
	scrollViewSelectMeasure.SetVisible(false)
	labelCalcErr.SetVisible(false)
	scrollViewCheckConsumption.SetVisible(true)

	runWork(func(ctx context.Context) error {

		defer func() {
			tableViewMeasure.SetVisible(true)
			scrollViewSelectMeasure.SetVisible(true)
			scrollViewCheckConsumption.SetVisible(false)
		}()

		for {
			select {
			case <-ctx.Done():
				return nil

			case gas := <-chGas:
				if err := switchGas(log, ctx, gas); err != nil {
					return err
				}
				appWindow.Synchronize(func() {
					_ = labelGas.SetText(fmt.Sprintf("%d", gas))
				})
			default:
				cons, err := readGasConsumption(log, ctx)
				if err != nil {
					return err
				}
				appWindow.Synchronize(func() {
					_ = labelConsumption.SetText(pkg.FormatFloatTrimNulls(cons, 3))
				})
				pause(ctx, 50*time.Millisecond)
			}
		}
	})
}

func errNeedBytesCount(what string, n, len int) error {
	return fmt.Errorf("%s: ожидалось %d байт ответа, получено %d", what, n, len)
}
