package app

import (
	"context"
	"errors"
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

	if gas < 0 {
		log.Panicln("gas: negative value", gas)
	}

	if gas == 0 {
		return switchOffGas(log, ctx)
	}

	setStatusOkSync(labelGasBlock, fmt.Sprintf("переключение %d", gas))
	b := []byte{0x06, 0x03, 0x03, 0x02, byte(gas), 0}
	addSumBytesToEnd(b)
	c := cfg.Get()

	log = structloge.PrependSuffixKeys(log, "COMPORT", c.Gas.Comport)

	r, err := c.CommGas().GetResponse(log, ctx, b)
	if err != nil {
		setStatusErrSync(labelGasBlock, err)
		return fmt.Errorf("газовый блок: %s: %w", c.Gas.Comm.Comport, err)
	}

	b = []byte{0x07, 0x03, 0x03, 0x02, byte(gas), 0x00, 0x00}
	addSumBytesToEnd(b)
	if !cmp.Equal(b, r) {
		err := fmt.Errorf("получен ответ % X, ожидалось % X", r, b)
		setStatusErrSync(labelGasBlock, err)
		return fmt.Errorf("газовый блок: %w", err)
	}
	setStatusOkSync(labelGasBlock, fmt.Sprintf("газ %d", gas))
	return nil
}

func addSumBytesToEnd(b []byte) {
	n := len(b) - 1
	b[n] = 0
	for i := range b[:n] {
		b[n] += b[i]
	}
}

func checkSumBytesEnd(b []byte) bool {
	n := len(b) - 1
	var x byte
	for i := range b[:n] {
		x += b[i]
	}
	return x == b[n]
}

func readGasConsumption(log comm.Logger, ctx context.Context) (float64, error) {
	c := cfg.Get()
	log = structloge.PrependSuffixKeys(log,
		"COMPORT", c.Gas.Comport,
		"gas_block_request", "`запрос расхода`")

	b := []byte{0x06, 0x03, 0x03, 0x04, c.Gas.ConsChan, 0}
	addSumBytesToEnd(b)

	r, err := c.CommGas().GetResponse(log, ctx, b)
	if err != nil {
		setStatusErrSync(labelGasBlock, err)
		return 0, fmt.Errorf("газовый блок: %s: %w", c.Gas.Comm.Comport, err)
	}
	if len(r) < 11 {
		return 0, errors.New("ожидалось 11 байт")
	}
	if !checkSumBytesEnd(r) {
		return 0, errors.New("не совпадает контрольная сумма")
	}
	bits := c.Gas.GetConsByteOrder().Uint32(r[6:])
	f32 := math.Float32frombits(bits)
	str := strconv.FormatFloat(float64(f32), 'f', -1, 32)
	f64, _ := strconv.ParseFloat(str, 64)
	if math.IsNaN(f64) || math.IsInf(f64, -1) || math.IsInf(f64, 1) || math.IsInf(f64, 0) {
		return f64, fmt.Errorf("not a float %v number", c.Gas.GetConsByteOrder())
	}
	q := -2.0*f64/1000. + 0.02
	a := 1.
	if q >= 0 {
		a = -1
	}
	q = (q + a) / 0.8
	setStatusOkSync(labelGasBlock, fmt.Sprintf("расход %v", q))
	return q, nil
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

func setupPlaceConnection(log comm.Logger, ctx context.Context, placeConnection uint16) error {
	setStatusOkSync(labelControlSheet, fmt.Sprintf("установка реле %016b", placeConnection))
	c := cfg.Get()
	log = structloge.PrependSuffixKeys(log, "COMPORT", c.ControlSheet.Comport)

	_, err := modbus.Request{
		Addr:     1,
		ProtoCmd: 0x10,
		Data:     []byte{0x00, 0x50, 0x00, 0x01, 0x02, byte(placeConnection >> 8), byte(placeConnection)},
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

	for i := 0; i < 16; i++ {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		// замкнуть место i, остальные места разомкнуть
		if err := setupPlaceConnection(log, ctx, 1<<i); err != nil {
			return err
		}

		// измерить напряжение
		if err := readVoltmeter(log, ctx, smp); err != nil {
			return err
		}

		// обрыв, если измеренное напряжение на месте i больше 6
		smp.Br[i] = math.Abs(smp.U[i]) > 6
	}

	var placeConnection uint16
	for i := 0; i < 16; i++ {
		if !smp.Br[i] {
			placeConnection |= 1 << i
		}
	}
	if err := setupPlaceConnection(log, ctx, placeConnection); err != nil {
		return err
	}

	setStatusOkSync(labelControlSheet, fmt.Sprintf("поиск обрыва: %016b", placeConnection))

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
		err := fmt.Errorf("%s: позиция %d: %w", string(b), 17, err)
		setStatusErrSync(labelVoltmeter, err)
		return fmt.Errorf("вольтметр: %w", err)
	}
	smp.I = math.Abs(Ui) / 30.1

	Ut, err := strconv.ParseFloat(xsStr[16], 64)
	if err != nil {
		err := fmt.Errorf("%s: позиция %d: %w", string(b), 16, err)
		setStatusErrSync(labelVoltmeter, err)
		return fmt.Errorf("вольтметр: %w", err)
	}
	smp.T = 8.969*math.Abs(Ut) - 64.305

	setStatusOkSync(labelVoltmeter, s)
	return nil
}
