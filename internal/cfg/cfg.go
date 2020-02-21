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
}

type Gas struct {
	Addr byte `yaml:"addr"`
	Comm `yaml:"comm"`
}

type Voltmeter struct {
	Comm `yaml:"comm"`
}

type Comm struct {
	BaudRate           int           `yaml:"baud_rate"`
	Comport            string        `yaml:"comport"`
	TimeoutGetResponse time.Duration `yaml:"timeout_get_response"`
	TimeoutEndResponse time.Duration `yaml:"timeout_end_response"`
	MaxAttemptsRead    int           `yaml:"max_attempts_read"`
}

func SetYaml(strYaml []byte) error {
	var c Config
	if err := yaml.Unmarshal(strYaml, &c); err != nil {
		return err
	}
	comm.SetEnableLog(c.LogComm)
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
	comm.SetEnableLog(c.LogComm)
	cfg = c
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
	}
	must.PanicIf(Set(c))
}

var (
	mu  sync.Mutex
	cfg Config
)
