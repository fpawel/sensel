package logfile

import (
	"fmt"
	"github.com/fpawel/sensel/internal/pkg/must"
	"os"
	"path/filepath"
	"time"
)

func MustNew(filenameSuffix string) *os.File {
	file, err := New(filenameSuffix)
	must.PanicIf(err)
	return file
}

func New(filenameSuffix string) (*os.File, error) {
	if err := ensureDir(); err != nil {
		return nil, err
	}
	filename := filename(daytime(time.Now()), filenameSuffix)
	return os.OpenFile(filename, os.O_CREATE|os.O_APPEND, 0666)
}

func filename(t time.Time, suffix string) string {
	return filepath.Join(logDir, fmt.Sprintf("%s%s.log", t.Format("2006-01-02"), suffix))
}

func daytime(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local)
}

func ensureDir() error {
	_, err := os.Stat(logDir)
	if os.IsNotExist(err) { // создать каталог если его нет
		err = os.MkdirAll(logDir, os.ModePerm)
	}
	return err
}

var (
	logDir = filepath.Join(filepath.Dir(os.Args[0]), "logs")
)
