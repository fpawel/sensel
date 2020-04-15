package app

import (
	"context"
	"fmt"
	"github.com/fpawel/comm"
	"github.com/fpawel/comm/modbus"
	"github.com/fpawel/sensel/internal/cfg"
	"github.com/fpawel/sensel/internal/data"
	"github.com/fpawel/sensel/internal/pkg/structloge"
	"github.com/google/go-cmp/cmp"
	"math"
	"strconv"
	"strings"
)

func switchOffGas(log comm.Logger, ctx context.Context) error {
	setStatusOkSync(labelGasBlock, "отключение")
	c := cfg.Get()

	log = structloge.PrependSuffixKeys(log, "COMPORT", c.Gas.Comport)

	r, err := c.CommGas().GetResponse(log, ctx, []byte{0x05, 0x03, 0x03, 0x03, 0x0E})
	if err != nil {
		setStatusErrSync(labelGasBlock, err)
		return fmt.Errorf("газовый блок: %s: %w", c.Gas.Comm.Comport, err)
	}
	a := []byte{0x06, 0x03, 0x03, 0x03, 0x00, 0x0F}
	if !cmp.Equal(a, r) {
		err := fmt.Errorf("получен ответ % X, ожидалось % X", r, a)
		setStatusErrSync(labelGasBlock, err)
		return fmt.Errorf("газовый блок: %w", err)
	}
	setStatusOkSync(labelGasBlock, "отключен")
	return nil
}

func switchGas(log comm.Logger, ctx context.Context, gas int) error {
	setStatusOkSync(labelGasBlock, fmt.Sprintf("переключение %d", gas))
	b := []byte{0x06, 0x03, 0x03, 0x02, byte(gas), 0}
	for i := range b[:len(b)-1] {
		b[5] += b[i]
	}
	c := cfg.Get()

	log = structloge.PrependSuffixKeys(log, "COMPORT", c.Gas.Comport)

	r, err := c.CommGas().GetResponse(log, ctx, b)
	if err != nil {
		setStatusErrSync(labelGasBlock, err)
		return fmt.Errorf("газовый блок: %s: %w", c.Gas.Comm.Comport, err)
	}

	a := []byte{0x07, 0x03, 0x03, 0x02, byte(gas), 0x00, 0x00}
	for i := range a[:len(a)-1] {
		a[6] += a[i]
	}
	if !cmp.Equal(a, r) {
		err := fmt.Errorf("получен ответ % X, ожидалось % X", r, a)
		setStatusErrSync(labelGasBlock, err)
		return fmt.Errorf("газовый блок: %w", err)
	}
	setStatusOkSync(labelGasBlock, fmt.Sprintf("газ %d", gas))
	return nil
}

func setupTensionBar(log comm.Logger, ctx context.Context, U float64) error {

	setStatusOkSync(labelControlSheet, fmt.Sprintf("установка напряжения %v", U))

	c := cfg.Get()
	v := uint16(math.Round(U))

	log = structloge.PrependSuffixKeys(log, "COMPORT", c.ControlSheet.Comport)

	b, err := modbus.Request{
		Addr:     1,
		ProtoCmd: 0x10,
		Data:     []byte{0x00, 0x30, 0x00, 0x01, 0x02, byte(v >> 8), byte(v)},
	}.GetResponse(log, ctx, c.CommControl())
	if err != nil {
		setStatusErrSync(labelControlSheet, err)
		return fmt.Errorf("стенд: %s: %w", c.ControlSheet.Comm.Comport, err)
	}
	if len(b) != 8 {
		err := errNeedBytesCount("стенд", 8, len(b))
		setStatusErrSync(labelControlSheet, err)
		return err
	}
	if b[3] != 0x30 {
		err := fmt.Errorf("стенд: 3-ий байт ответа: %d, ожидалось 0x30", b[3])
		setStatusErrSync(labelControlSheet, err)
		return err
	}

	setStatusOkSync(labelControlSheet, fmt.Sprintf("установлено напряжение %v", U))

	return nil
}

type placeConnection int

const (
	disconnect placeConnection = iota
	connect
)

func (x placeConnection) String() string {
	if x == connect {
		return "ВКЛ"
	}
	return "ВЫКЛ"
}

func (x placeConnection) byte() byte {
	if x == connect {
		return 0
	}
	return 1
}

func setupPlaceConnection(log comm.Logger, ctx context.Context, placeConnection placeConnection, place int) error {
	setStatusOkSync(labelControlSheet, fmt.Sprintf("установка реле %s %d", placeConnection, place))
	c := cfg.Get()
	log = structloge.PrependSuffixKeys(log, "COMPORT", c.ControlSheet.Comport)

	_, err := modbus.Request{
		Addr:     1,
		ProtoCmd: 0x10,
		Data:     []byte{0x00, 0x50, 0x00, 0x01, 0x02, placeConnection.byte(), byte(place)},
	}.GetResponse(log, ctx, c.CommControl())
	if err != nil {
		setStatusErrSync(labelControlSheet, err)
		return fmt.Errorf("стенд: %s: %w", c.ControlSheet.Comm.Comport, err)
	}
	return nil
}

func setupCurrentBar(log comm.Logger, ctx context.Context, I float64) error {

	setStatusOkSync(labelControlSheet, fmt.Sprintf("установка тока %v", I))

	c := cfg.Get()
	v := uint16(math.Round(I / c.ControlSheet.KI))

	log = structloge.PrependSuffixKeys(log, "COMPORT", c.ControlSheet.Comport)

	b, err := modbus.Request{
		Addr:     1,
		ProtoCmd: 0x10,
		Data:     []byte{0x00, 0x20, 0x00, 0x01, 0x02, byte(v >> 8), byte(v)},
	}.GetResponse(log, ctx, c.CommControl())
	if err != nil {
		setStatusErrSync(labelControlSheet, err)
		return fmt.Errorf("стенд: %s: %w", c.ControlSheet.Comm.Comport, err)
	}

	if len(b) != 8 {
		err := errNeedBytesCount("стенд", 8, len(b))
		setStatusErrSync(labelControlSheet, err)
		return err
	}
	if b[3] != 0x20 {
		err := fmt.Errorf("стенд: 3-ий байт ответа: %d, ожидалось 0x20", b[3])
		setStatusErrSync(labelControlSheet, err)
		return err
	}

	setStatusOkSync(labelControlSheet, fmt.Sprintf("установлен ток %v", I))

	return nil
}

func readBreak(log comm.Logger, ctx context.Context, smp *data.Sample) error {

	setStatusOkSync(labelControlSheet, "поиск обрыва")

	c := cfg.Get()

	log = structloge.PrependSuffixKeys(log, "COMPORT", c.ControlSheet.Comport)

	// установить напряжение питания 10 В
	if err := setupTensionBar(log, ctx, 10); err != nil {
		return err
	}
	for i := 0; i < 16; i++ {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		// закоротить место i, остальные места разомкнуть
		if err := setupPlaceConnection(log, ctx, connect, i+1); err != nil {
			return err
		}
		// измерить напряжение
		if err := readVoltmeter(log, ctx, smp); err != nil {
			return err
		}
		// обрыв, если измеренное напряжение на месте i больше 6
		smp.Br[i] = math.Abs(smp.U[i]) > 6
	}
	// установить рабочее напряжение питания
	if err := setupTensionBar(log, ctx, smp.Ub); err != nil {
		return err
	}
	setStatusOkSync(labelControlSheet, fmt.Sprintf("поиск обрыва: %v", smp.Br))
	return nil
}

func readVoltmeter(log comm.Logger, ctx context.Context, smp *data.Sample) error {

	Conf := cfg.Get()
	cfgComm := Conf.Voltmeter.Comm.Comm()

	log = structloge.PrependSuffixKeys(log, "COMPORT", Conf.Voltmeter.Comport)

	Comport := Conf.ComportVoltmeter()

	const scanRequest1 = "ROUTe:SCAN:SCAN"
	const scanRequest = scanRequest1 + "\n"

	setStatusOkSync(labelVoltmeter, scanRequest1)

	err := comm.Write(ctx, []byte(scanRequest), Comport, cfgComm)
	if err != nil {
		err := fmt.Errorf("вольтметр: %s: %w", scanRequest1, err)
		setStatusErrSync(labelVoltmeter, err)
		return err
	}

	if Conf.Debug.LogComm {
		log.Printf("вольтметр: %s: % X", scanRequest1, []byte(scanRequest))
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
	for i := range smp.U {
		smp.U[i], err = strconv.ParseFloat(xsStr[i], 64)
		if err != nil {
			err := fmt.Errorf("%s: позиция %d: %w", string(b), i, err)
			setStatusErrSync(labelVoltmeter, err)
			return fmt.Errorf("вольтметр: %w", err)
		}
	}

	Ui, err := strconv.ParseFloat(xsStr[17], 64)
	if err != nil {
		err := fmt.Errorf("%s: позиция %d: %w", string(b), 16, err)
		setStatusErrSync(labelVoltmeter, err)
		return fmt.Errorf("вольтметр: %w", err)
	}
	smp.I = math.Abs(Ui) / 30.1
	setStatusOkSync(labelVoltmeter, s)
	return nil
}
