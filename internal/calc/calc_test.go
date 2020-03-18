package calc

import (
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
)

func TestNew(t *testing.T) {
	filename := filepath.Join(os.Getenv("GOPATH"), "src", "github.com", "fpawel", "sensel",
		"build", "lua", "sensel.lua")
	_, err := New(filename)
	assert.NoError(t, err)
}
