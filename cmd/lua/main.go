package main

import (
	"fmt"
	lua "github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"
)

type nameValueOk struct {
	Name  string
	Value float64
	Ok    bool
}

func column(name string, value float64, ok bool) nameValueOk {
	return nameValueOk{
		Name:  name,
		Value: value,
		Ok:    ok,
	}
}

type T struct {
	Name    string
	Columns map[string][]nameValueOk
}

func main() {
	L := lua.NewState()
	defer L.Close()

	var Config T

	L.SetGlobal("column", luar.New(L, column))
	L.SetGlobal("Config", luar.New(L, &Config))

	err := L.DoString(`
Config.Name = "Ибрагим"
Config.Columns = {
	["dd"] = {
		column("xx", 1, true),
		column("yy", 2, false),
		column("yy&&&", 332, false),
	},
	["cc"] = {
		column("AA", 21, true),
		column("BB", 32, false),
		{Name = "^^&&NNN", Value = 5555, Ok = false},
	},
}`)
	if err != nil {
		panic(err)
	}
	fmt.Println(Config.Name)
	fmt.Printf("%+v", Config.Columns)

}
