package app

import (
	"fmt"
	"github.com/fpawel/sensel/internal/data"
	. "github.com/lxn/walk/declarative"
	"math"
	"math/rand"
	"time"
)

func newApplicationWindow() MainWindow {

	measurementViewModel = &MeasurementViewModel{
		M: data.Measurement{
			Samples: randSamples(),
		},
	}

	return MainWindow{
		AssignTo: &appWindow,
		Title:    "ЧЭ лаборатория 74",
		Font: Font{
			Family:    "Segoe UI",
			PointSize: 10,
		},
		Layout: VBox{},
		MenuItems: []MenuItem{
			Menu{
				Text: "Опрос",
			},
			Menu{
				Text: "Обмер",
			},
			Menu{
				Text: "Прервать",
			},
			Menu{
				Text: "Настройки",
			},
			Menu{
				Text: "Конфигурация",
			},
			Menu{
				Text: "Сценарий",
			},
		},
		Children: []Widget{
			TableView{
				//AlternatingRowBG:         true,
				//DoubleBuffering:          true,
				Columns:                  measurementViewModel.Columns(),
				ColumnsOrderable:         false,
				ColumnsSizable:           true,
				LastColumnStretched:      false,
				Model:                    measurementViewModel,
				MultiSelection:           true,
				NotSortableByHeaderClick: true,
			},
		},
	}
}

func randSamples() []data.Sample {
	xs := make([]data.Sample, 10)
	for i := range xs {
		for j := 0; j < 16; j++ {
			xs[i].Productions = randProductions()
			xs[i].Temperature = rand3()
			xs[i].Current = rand3()
			xs[i].Consumption = rand3()
			xs[i].CreatedAt = time.Now()
			xs[i].Name = fmt.Sprintf("X%d", i)
		}
	}
	return xs
}

func randProductions() []data.Production {
	xs := make([]data.Production, 16)
	for i := range xs {
		xs[i].Place = i
		xs[i].Break = rand.Float64() < 0.1
		xs[i].Value = rand3()
	}
	return xs
}
func rand3() float64 {
	return math.Round(rand.Float64()*1000) / 1000
}

func init() {
	rand.NewSource(time.Now().UnixNano())
}
