package app

import (
	"context"
	"fmt"
	"github.com/fpawel/comm"
	"github.com/fpawel/comm/modbus"
	"github.com/fpawel/sensel/internal/calc"
	"github.com/fpawel/sensel/internal/cfg"
	"github.com/fpawel/sensel/internal/data"
	"github.com/fpawel/sensel/internal/pkg/must"
	"github.com/google/go-cmp/cmp"
	"math"
	"strconv"
	"strings"
	"time"
)

func runMeasure(measurementScheme []calc.Sample) {
	measurement := data.Measurement{
		MeasurementInfo: data.MeasurementInfo{
			MeasurementID: 0,
			CreatedAt:     time.Now(),
			Name:          lineEditMeasureName.Text(),
			Device:        comboBoxDevice.Text(),
			Kind:          comboBoxKind.Text(),
		},
		MeasurementData: data.MeasurementData{
			Pgs: []float64{
				numberEditC[0].Value(),
				numberEditC[1].Value(),
				numberEditC[2].Value(),
				numberEditC[3].Value(),
			},
		},
	}
	setMeasurement(measurement)
	runWork(func(ctx context.Context) error {
		measureStartTime := time.Now()
		for nSample, smp := range measurementScheme {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			measurement.Samples = append(measurement.Samples, data.Sample{
				Tm:  time.Now(),
				Gas: smp.Gas,
				Ub:  smp.Tension,
				I:   smp.Current,
			})

			appWindow.Synchronize(func() {
				s := fmt.Sprintf("Измерение %d газ=%d U=%g I=%g: установка рабочего режима",
					nSample+1, smp.Gas, smp.Tension, smp.Current)
				must.PanicIf(labelCurrentWork.SetText(s))
				labelCurrentWork.SetVisible(true)
			})

			if err := setupMeasurement(log, ctx, smp); err != nil {
				log.PrintErr(err)
				//return err
			}
			ctxDelay, _ := context.WithTimeout(ctx, smp.Duration)
			delay(log, ctxDelay, nSample, measurementScheme, measureStartTime, &measurement)
		}
		return nil
	})
}

func delay(log comm.Logger, ctx context.Context, smpIndex int, samples []calc.Sample, measureStartTime time.Time, m *data.Measurement) {

	progress := func(t time.Time, d time.Duration) int {
		return int(100 * float64(time.Since(t)) / float64(d))
	}

	smp := samples[smpIndex]
	tickUpdGui := time.NewTicker(time.Second * 2)
	tickReadSample := time.NewTicker(cfg.Get().ReadSampleInterval)
	startTime := time.Now()

	upd := func() {
		d1 := progress(startTime, smp.Duration)
		totalDur := calc.GetTotalMeasureDuration(samples)
		d2 := progress(measureStartTime, totalDur)
		progressBarCurrentWork.SetValue(d1)
		progressBarTotalWork.SetValue(d2)
		s := fmt.Sprintf("Измерение %d газ=%d U=%g I=%g, %s из %s %d%s",
			smpIndex+1, smp.Gas, smp.Tension, smp.Current, time.Since(startTime), smp.Duration,
			d1, "%")
		_ = labelCurrentWork.SetText(s)
		s = fmt.Sprintf("Общий прогресс выполнения %s из %s %d%s",
			time.Since(measureStartTime), totalDur, d2, "%")
		_ = labelTotalWork.SetText(s)
	}

	appWindow.Synchronize(func() {
		progressBarCurrentWork.SetVisible(true)
		progressBarTotalWork.SetVisible(true)

		progressBarCurrentWork.SetRange(0, 100)
		progressBarTotalWork.SetRange(0, 100)
		progressBarTotalWork.SetRange(0, 100)
		labelCurrentWork.SetVisible(true)
		labelTotalWork.SetVisible(true)
		upd()
	})

	defer func() {
		tickUpdGui.Stop()
		appWindow.Synchronize(func() {
			progressBarCurrentWork.SetVisible(false)
			progressBarTotalWork.SetVisible(false)
			labelCurrentWork.SetVisible(false)
			labelTotalWork.SetVisible(false)
		})
	}()
	for {
		select {
		case <-ctx.Done():
			return
		case <-tickUpdGui.C:
			appWindow.Synchronize(upd)
		case <-tickReadSample.C:
			dataSmp := &m.Samples[len(m.Samples)-1]
			_ = readBar(log, ctx, dataSmp)
			appWindow.Synchronize(func() {
				setMeasurement(*m)
			})
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
		return err
	}
	return nil
}

func runReadSample() {
	runWork(func(ctx context.Context) error {
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
		return nil
	})
}

func switchOffGas(log comm.Logger, ctx context.Context) error {
	c := cfg.Get()
	r, err := c.CommGas().GetResponse(log, ctx, []byte{0x05, 0x03, 0x03, 0x03, 0x0E})
	if err != nil {
		return fmt.Errorf("газовый блок: %s: %w", c.Gas.Comm.Comport, err)
	}
	a := []byte{0x06, 0x03, 0x03, 0x03, 0x00, 0x0F}
	if !cmp.Equal(a, r) {
		return fmt.Errorf("газовый блок: получен ответ % X, ожидалось % X", r, a)
	}
	return nil
}

func switchGas(log comm.Logger, ctx context.Context, gas int) error {
	b := []byte{0x06, 0x03, 0x03, 0x02, byte(gas), 0}
	for i := range b[:len(b)-1] {
		b[5] += b[i]
	}
	c := cfg.Get()
	r, err := c.CommGas().GetResponse(log, ctx, b)
	if err != nil {
		return fmt.Errorf("газовый блок: %s: %w", c.Gas.Comm.Comport, err)
	}

	a := []byte{0x07, 0x03, 0x03, 0x02, byte(gas), 0x00, 0x00}
	for i := range a[:len(a)-1] {
		a[6] += a[i]
	}
	if !cmp.Equal(a, r) {
		return fmt.Errorf("газовый блок: получен ответ % X, ожидалось % X", r, a)
	}
	return nil
}

func readBar(log comm.Logger, ctx context.Context, smp *data.Sample) error {
	if err := readBreak(log, ctx, smp); err != nil {
		log.PrintErr(err)
	}
	if err := readVoltmeter(log, ctx, smp.U[:]); err != nil {
		return err
	}
	smp.Tm = time.Now()
	return nil
}

func readBreak(log comm.Logger, ctx context.Context, smp *data.Sample) error {
	c := cfg.Get()
	b, err := modbus.Request{
		Addr:     1,
		ProtoCmd: 0x03,
		Data:     []byte{0x00, 0x40, 0x00, 0x01},
	}.GetResponse(log, ctx, c.CommControl())
	if err != nil {
		return fmt.Errorf("стенд: поиск обрыва: %w", err)
	}
	if len(b) != 7 {
		return fmt.Errorf("стенд: поиск обрыва: ожидалось 7 байт ответа, получено %d", len(b))
	}
	for i := 0; i < 8; i++ {
		smp.Br[i] = b[4]&(1<<i) != 0
		smp.Br[i+8] = b[5]&(1<<i) != 0
	}
	return nil
}

func readVoltmeter(log comm.Logger, ctx context.Context, result []float64) error {
	Conf := cfg.Get()
	cfgComm := Conf.Voltmeter.Comm.Comm()
	Comport := Conf.ComportVoltmeter()
	err := comm.Write(ctx, []byte("ROUTe:SCAN:SCAN\n"), Comport, cfgComm)
	if err != nil {
		return fmt.Errorf("вольтметр: ROUTe:SCAN:SCAN: %w", err)
	} else {
		if Conf.Debug.LogComm {
			log.Printf("вольтметр: ROUTe:SCAN:SCAN: % X", []byte("ROUTe:SCAN:SCAN\n"))
		}
	}

	pause(ctx.Done(), Conf.Voltmeter.PauseScan)
	b, err := Conf.CommVoltmeter().GetResponse(log, ctx, []byte("FETCh?\n"))
	if err != nil {
		return fmt.Errorf("вольтметр: %w", err)
	}
	s := string(b)
	s = strings.TrimSpace(s)
	xsStr := strings.Split(s, ",")
	if len(xsStr) != 20 {
		return fmt.Errorf("вольтметр: ожидалось 20 разделённых запятой значений: %s", string(b))
	}
	for i := range xsStr {
		if i < len(result) {
			result[i], err = strconv.ParseFloat(xsStr[i], 64)
			if err != nil {
				return fmt.Errorf("вольтметр: %s: позиция %d: %w", string(b), i, err)
			}
		}
	}
	return nil
}

func setupTensionBar(log comm.Logger, ctx context.Context, U float64) error {
	c := cfg.Get()
	v := uint16(math.Round(U))

	b, err := modbus.Request{
		Addr:     1,
		ProtoCmd: 0x10,
		Data:     []byte{0x00, 0x30, 0x00, 0x01, 0x02, byte(v >> 8), byte(v)},
	}.GetResponse(log, ctx, c.CommControl())
	if err != nil {
		return fmt.Errorf("стенд: %s: %w", c.Control.Comm.Comport, err)
	}
	if len(b) != 8 {
		return errNeedBytesCount("стенд", 8, len(b))
	}
	if b[3] != 0x30 {
		return fmt.Errorf("стенд: 3-ий байт ответа: %d, ожидалось 0x30", b[3])
	}
	return nil
}

func setupCurrentBar(log comm.Logger, ctx context.Context, I float64) error {
	c := cfg.Get()
	v := uint16(math.Round(I / c.Control.KI))

	b, err := modbus.Request{
		Addr:     1,
		ProtoCmd: 0x10,
		Data:     []byte{0x00, 0x20, 0x00, 0x01, 0x02, byte(v >> 8), byte(v)},
	}.GetResponse(log, ctx, c.CommControl())
	if err != nil {
		return fmt.Errorf("стенд: %s: %w", c.Control.Comm.Comport, err)
	}

	if len(b) != 8 {
		return errNeedBytesCount("стенд", 8, len(b))
	}
	if b[3] != 0x20 {
		return fmt.Errorf("стенд: 3-ий байт ответа: %d, ожидалось 0x20", b[3])
	}
	return nil
}

func pause(chDone <-chan struct{}, d time.Duration) {
	timer := time.NewTimer(d)
	for {
		select {
		case <-timer.C:
			return
		case <-chDone:
			timer.Stop()
			return
		}
	}
}

func errNeedBytesCount(what string, n, len int) error {
	return fmt.Errorf("%s: ожидалось %d байт ответа, получено %d", what, n, len)
}
