package cfg

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/fpawel/comm"
	"github.com/fpawel/sensel/internal/pkg/must"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Config struct {
	Printer      string    `yaml:"printer"`
	Gas          Gas       `yaml:"gas"`
	Voltmeter    Voltmeter `yaml:"voltmeter"`
	ControlSheet Control   `yaml:"control"`
	Debug        struct {
		LogComm bool `yaml:"log_comm"`
	} `yaml:"debug"`
	Table                 TableConfig `yaml:"table"`
	LastMeasurementsCount int         `yaml:"last_measurements_count"`
	AppWindow             AppWindow   `yaml:"app_window"`
}

type Gas struct {
	Comm `yaml:"comm"`
}

type Voltmeter struct {
	Comm             `yaml:"comm"`
	PauseScan        time.Duration `yaml:"pause_scan"`
	PauseMeasureScan time.Duration `yaml:"pause_measure_scan"`
}

type Control struct {
	Comm `yaml:"comm"`
	KI   float64 `yaml:"Ki"`
	Kt0  float64 `yaml:"Kt0"`
	Kt1  float64 `yaml:"Kt1"`
}

type Comm struct {
	BaudRate           int           `yaml:"baud_rate"`
	Comport            string        `yaml:"comport"`
	TimeoutGetResponse time.Duration `yaml:"timeout_get_response"`
	TimeoutEndResponse time.Duration `yaml:"timeout_end_response"`
	MaxAttemptsRead    int           `yaml:"max_attempts_read"`
}

type TableConfig struct {
	RowHeightMM      float64 `yaml:"row_height_mm"`
	CellHorizSpaceMM float64 `yaml:"cell_horiz_space_mm"`
	FontSizePixels   float64 `yaml:"font_size_pixels"`
	IncludeSamples   bool    `yaml:"include_samples"`
}

type AppWindow struct {
	TableViewMeasure TableViewMeasure `yaml:"table_view_measure"`
}

type TableViewMeasure struct {
	ColumnWidths []int `yaml:"column_widths"`
}

func (x Comm) Comm() comm.Config {
	return comm.Config{
		TimeoutGetResponse: x.TimeoutGetResponse,
		TimeoutEndResponse: x.TimeoutEndResponse,
		MaxAttemptsRead:    x.MaxAttemptsRead,
	}
}

func Get() (r Config) {
	mu.Lock()
	defer mu.Unlock()
	return getGob()
}

func Set(c Config) error {

	if c.Voltmeter.PauseMeasureScan < time.Second {
		c.Voltmeter.PauseMeasureScan = time.Second
	}

	b := must.MarshalYaml(c)
	mu.Lock()
	defer mu.Unlock()
	if err := writeFile(b); err != nil {
		return err
	}
	cfg = c
	comm.SetEnableLog(c.Debug.LogComm)
	return nil
}

func writeFile(b []byte) error {
	return ioutil.WriteFile(filename(), b, 0666)
}

func filename() string {
	return filepath.Join(filepath.Dir(os.Args[0]), "config.yaml")
}

func readFile() (Config, error) {
	var c Config
	data, err := ioutil.ReadFile(filename())
	if err != nil {
		return c, err
	}
	err = yaml.Unmarshal(data, &c)
	return c, err
}

func init() {
	c, err := readFile()
	defer func() {
		must.PanicIf(Set(c))
	}()
	if err == nil {
		return
	}
	fmt.Println(err, "file:", filename())

	c = Config{
		Gas: Gas{
			Comm: Comm{
				BaudRate:           9600,
				TimeoutGetResponse: time.Second,
				TimeoutEndResponse: 50 * time.Millisecond,
				MaxAttemptsRead:    3,
			},
		},
		Voltmeter: Voltmeter{
			Comm: Comm{
				BaudRate:           115200,
				TimeoutGetResponse: time.Second,
				TimeoutEndResponse: 50 * time.Millisecond,
				MaxAttemptsRead:    3,
			},
			PauseScan:        3 * time.Second,
			PauseMeasureScan: time.Second,
		},
		ControlSheet: Control{
			Comm: Comm{
				BaudRate:           9600,
				TimeoutGetResponse: 3 * time.Second,
				TimeoutEndResponse: 50 * time.Millisecond,
				MaxAttemptsRead:    3,
			},
			KI:  0.000082,
			Kt0: -64.305,
			Kt1: 8.969,
		},
		Table: TableConfig{
			RowHeightMM:      2.75,
			CellHorizSpaceMM: 1.,
			FontSizePixels:   5.5,
		},
		LastMeasurementsCount: 50,
	}
}

func getGob() (r Config) {
	must.PanicIf(enc.Encode(cfg))
	must.PanicIf(dec.Decode(&r))
	return
}

var (
	mu  sync.Mutex
	cfg Config

	buff = new(bytes.Buffer)
	enc  = gob.NewEncoder(buff)
	dec  = gob.NewDecoder(buff)
)
