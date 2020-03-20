package main

import (
	"github.com/fpawel/sensel/internal/app"
	"github.com/powerman/structlog"
	"os"
	"path/filepath"
)

func main() {

	structlog.DefaultLogger.
		SetPrefixKeys(structlog.KeyLevel).
		SetSuffixKeys(structlog.KeyUnit, structlog.KeySource, structlog.KeyStack).
		SetDefaultKeyvals(
			structlog.KeyApp, filepath.Base(os.Args[0]),
			structlog.KeySource, structlog.Auto,
			//structlog.KeyStack, structlog.Auto,
		).
		SetKeysFormat(map[string]string{
			structlog.KeyTime:   " %[2]s",
			structlog.KeySource: " %6[2]s",
			structlog.KeyUnit:   " %6[2]s",
		})
	app.Main()
}
