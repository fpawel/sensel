package app

import (
	"github.com/fpawel/comm/comport"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"sort"
)

func ComboBoxWithList(values []string, getFunc func() string, setFunc func(string)) ComboBox {
	var cb *walk.ComboBox
	n := -1
	for i, s := range values {
		if s == getFunc() {
			n = i
			break
		}
	}
	return ComboBox{
		Editable:     true,
		AssignTo:     &cb,
		MaxSize:      Size{100, 0},
		Model:        values,
		CurrentIndex: n,
		OnCurrentIndexChanged: func() {
			setFunc(cb.Text())
		},
	}
}

func ComboBoxComport(getComportNameFunc func() string, setComportNameFunc func(string)) ComboBox {
	ports, _ := comport.Ports()
	sort.Strings(ports)
	return ComboBoxWithList(ports, getComportNameFunc, setComportNameFunc)
}
