package cfg

import (
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
	LogComm   bool      `yaml:"log_comm"`
	Gas       Gas       `yaml:"gas"`
	Voltmeter Voltmeter `yaml:"voltmeter"`
	Control   Control   `yaml:"control"`
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
}

type Comm struct {
	BaudRate           int           `yaml:"baud_rate"`
	Comport            string        `yaml:"comport"`
	TimeoutGetResponse time.Duration `yaml:"timeout_get_response"`
	TimeoutEndResponse time.Duration `yaml:"timeout_end_response"`
	MaxAttemptsRead    int           `yaml:"max_attempts_read"`
}

func (x Comm) Comm() comm.Config {
	return comm.Config{
		TimeoutGetResponse: x.TimeoutGetResponse,
		TimeoutEndResponse: x.TimeoutEndResponse,
		MaxAttemptsRead:    x.MaxAttemptsRead,
	}
}

func SetYaml(strYaml []byte) error {
	var c Config
	if err := yaml.Unmarshal(strYaml, &c); err != nil {
		return err
	}
	mu.Lock()
	defer mu.Unlock()
	must.PanicIf(writeFile(strYaml))
	cfg = c
	return nil
}

func Get() (r Config) {
	mu.Lock()
	defer mu.Unlock()
	must.UnmarshalJson(must.MarshalJson(cfg), &r)
	return
}

func Set(c Config) error {
	b := must.MarshalYaml(c)
	mu.Lock()
	defer mu.Unlock()
	if err := writeFile(b); err != nil {
		return err
	}
	cfg = c
	comm.SetEnableLog(c.LogComm)
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
			Control: Control{
				Comm{
					BaudRate:           9600,
					TimeoutGetResponse: time.Second,
					TimeoutEndResponse: 50 * time.Millisecond,
					MaxAttemptsRead:    3,
				},
			},
		}

	}
	must.PanicIf(Set(c))
}

var (
	mu  sync.Mutex
	cfg Config
)
