package app

import (
	"context"
	"github.com/ansel1/merry"
	lua "github.com/yuin/gopher-lua"
	"strconv"
	"time"
)

type luaConsole struct {
	L *lua.LState
}

func (x *luaConsole) errorIf(err error) {
	if err == nil || merry.Is(err, context.Canceled) {
		return
	}
	x.L.RaiseError("%v", err)
}

func (x *luaConsole) Pause(strDur string) {
	dur, err := time.ParseDuration(strDur)
	x.errorIf(err)
	pause(x.L.Context(), dur)
}

func (x *luaConsole) Gas(gas int) {
	x.errorIf(switchGas(log, x.L.Context(), gas))
}

func (x *luaConsole) SetTension(tension float64) {
	x.errorIf(setupTensionBar(log, x.L.Context(), tension))
}

func (x *luaConsole) SetCurrent(current float64) {
	x.errorIf(setupCurrentBar(log, x.L.Context(), current))
}

func (x *luaConsole) SetConnection(placeConnection uint16) {
	x.errorIf(setupPlaceConnection(log, x.L.Context(), placeConnection))
}

func (x *luaConsole) SetConnectionB(strPlaceConnection string) {
	v, err := strconv.ParseInt(strPlaceConnection, 2, 17)
	if err != nil {
		x.L.ArgError(2, err.Error())
	}
	x.errorIf(setupPlaceConnection(log, x.L.Context(), uint16(v)))
}

func (x *luaConsole) ReadGasConsumption() {
	_, err := readGasConsumption(log, x.L.Context())
	x.errorIf(err)
}
