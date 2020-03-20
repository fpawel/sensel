package app

import (
	"context"
	"fmt"
	"github.com/fpawel/comm"
	"github.com/fpawel/sensel/internal/cfg"
	"github.com/fpawel/sensel/internal/data"
	"github.com/fpawel/sensel/internal/pkg/comports"
	"strconv"
	"strings"
	"time"
)

func doReadSample() {
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

func readBar(log comm.Logger, ctx context.Context) (data.Sample, error) {
	smp := data.Sample{Tm: time.Now()}
	if err := readVoltmeter(log, ctx, smp.U[:]); err != nil {
		return smp, err
	}
	return smp, nil
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
