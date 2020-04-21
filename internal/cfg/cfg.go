package cfg

import (
	"fmt"
	"github.com/fpawel/comm"
	"github.com/fpawel/comm/comport"
	"github.com/fpawel/sensel/internal/pkg/comports"
	"github.com/fpawel/sensel/internal/pkg/must"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Config struct {
	Gas                Gas           `yaml:"gas"`
	Voltmeter          Voltmeter     `yaml:"voltmeter"`
	ControlSheet       Control       `yaml:"control"`
	ReadSampleInterval time.Duration `yaml:"read_sample_interval"`
	Debug              struct {
		LogComm bool `yaml:"log_comm"`
	} `yaml:"debug"`

	Table                 TableConfig `yaml:"table"`
	LastMeasurementsCount int         `yaml:"last_measurements_count"`
}

type Gas struct {
	Addr byte `yaml:"addr"`
	Comm `yaml:"comm"`
}

type Voltmeter struct {
	Comm      `yaml:"comm"`
	PauseScan time.Duration `yaml:"pause_scan"`
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

func (x Config) CommControl() comm.T {
	c := x.ControlSheet
	return comm.New(comports.GetComport(c.Comport, c.BaudRate), c.Comm.Comm())
}

func (x Config) CommGas() comm.T {
	c := x.Gas
	return comm.New(comports.GetComport(c.Comport, c.BaudRate), c.Comm.Comm())
}

func (x Config) CommVoltmeter() comm.T {
	return comm.New(x.ComportVoltmeter(), x.Voltmeter.Comm.Comm())
}

func (x Config) ComportVoltmeter() *comport.Port {
	c := x.Voltmeter
	return comports.GetComport(c.Comport, c.BaudRate)
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
	must.UnmarshalJson(must.MarshalJson(cfg), &r)
	return
}

func Set(c Config) error {

	if c.ReadSampleInterval < 10*time.Second {
		c.ReadSampleInterval = 10 * time.Second
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
	var err error
	c, err := readFile()
	if err != nil {
		fmt.Println(err, "file:", filename())

		c = Config{
			ReadSampleInterval: 5 * time.Second,
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
				PauseScan: 3 * time.Second,
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
	must.PanicIf(Set(c))
}

var (
	mu  sync.Mutex
	cfg Config
)
