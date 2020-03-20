package app

import (
	"context"
	"fmt"
	"github.com/fpawel/comm"
	"github.com/fpawel/comm/modbus"
	"github.com/fpawel/sensel/internal/cfg"
	"github.com/fpawel/sensel/internal/data"
	"github.com/fpawel/sensel/internal/pkg/comports"
	"math"
	"strconv"
	"strings"
	"time"
)

func runMeasure() error {
	measurementScheme, err := Calc.GetProductTypeMeasurementScheme(comboBoxDevice.Text(), comboBoxKind.Text())
	if err != nil {
		return err
	}

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
		for _, sampleScheme := range measurementScheme {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			if err := setupCurrentBar(log, ctx, sampleScheme.Current); err != nil {
				return err
			}
			if err := setupTensionBar(log, ctx, sampleScheme.Tension); err != nil {
				return err
			}

		}
		return nil
	})
	return nil
}

func runReadSample() {
	runWork(func(ctx context.Context) error {
		smp, err := readBar(log, ctx)
		if err != nil {
			return err
		}
		d := getMeasureTableViewModel().ViewData()
		d.Samples = []data.Sample{smp}
		appWindow.Synchronize(func() {
			labelCalcErr.SetVisible(false)
			getMeasureTableViewModel().SetViewData(d, nil)
		})
		return nil
	})
}

func switchGas(log comm.Logger, ctx context.Context, gas int) error {

}

func readBar(log comm.Logger, ctx context.Context) (data.Sample, error) {
	smp := data.Sample{Tm: time.Now()}
	c := cfg.Get()
	if err := readBreak(log, ctx, &smp); err != nil {
		if c.Debug.IgnoreReadBreakError {
			log.PrintErr(err)
		} else {
			return smp, err
		}
	}
	if err := readVoltmeter(log, ctx, smp.U[:]); err != nil {
		return smp, err
	}
	return smp, nil
}

func readBreak(log comm.Logger, ctx context.Context, smp *data.Sample) error {
	c := cfg.Get()
	b, err := modbus.Request{
		Addr:     1,
		ProtoCmd: 0x03,
		Data:     []byte{0x00, 0x40, 0x00, 0x01},
	}.GetResponse(log, ctx, c.CommControl())
	if err != nil {
		return fmt.Errorf("поиск обрыва: %w", err)
	}
	if len(b) != 7 {
		return fmt.Errorf("поиск обрыва: ожидалось 7 байт ответа, получено %d", len(b))
	}
	for i := 0; i < 8; i++ {
		smp.Br[i] = b[4]&(1<<i) != 0
		smp.Br[i+8] = b[5]&(1<<i) != 0
	}
	return nil
}

func readVoltmeter(log comm.Logger, ctx context.Context, result []float64) error {
	c := cfg.Get().Voltmeter
	cfgComm := c.Comm.Comm()
	cm := comports.GetComport(c.Comport, c.BaudRate)
	err := comm.Write(ctx, []byte("ROUTe:SCAN:SCAN\n"), cm, cfgComm)
	if err != nil {
		return err
	}
	pause(ctx.Done(), c.PauseScan)
	b, err := comm.New(cm, cfgComm).GetResponse(log, ctx, []byte("FETCh?\n"))
	if err != nil {
		return err
	}
	s := string(b)
	s = strings.TrimSpace(s)
	xsStr := strings.Split(s, ",")
	if len(xsStr) != 20 {
		return fmt.Errorf("ожидалось 20 разделённых запятой значений: %s", string(b))
	}
	for i := range xsStr {
		if i < len(result) {
			result[i], err = strconv.ParseFloat(xsStr[i], 64)
			if err != nil {
				return fmt.Errorf("%s: позиция %d: %w", string(b), i, err)
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
		return err
	}
	if len(b) != 8 {
		return errNeedBytesCount(8, len(b))
	}
	if b[3] != 0x30 {
		return fmt.Errorf("3-ий байт ответа:%d, ожидалось 0x30", b[3])
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
		return err
	}

	if len(b) != 8 {
		return errNeedBytesCount(8, len(b))
	}
	if b[3] != 0x20 {
		return fmt.Errorf("3-ий байт ответа:%d, ожидалось 0x20", b[3])
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

func errNeedBytesCount(n, len int) error {
	return fmt.Errorf("ожидалось %d байт ответа, получено %d", n, len)
}
