package must

import (
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

// FatalIf will call fmt.Print(err) and os.Exit(1) in case given err is not nil.
func FatalIf(err error) {
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}
}

// PanicIf will call panic(err) in case given err is not nil.
func PanicIf(err error) {
	if err != nil {
		panic(err)
	}
}

// MarshalYaml is a wrapper for toml.Marshal.
func MarshalYaml(v interface{}) []byte {
	data, err := yaml.Marshal(v)
	PanicIf(err)
	return data
}

func MarshalJsonIndent(v interface{}, prefix, indent string) []byte {
	data, err := json.MarshalIndent(v, prefix, indent)
	PanicIf(err)
	return data
}

func MarshalJson(v interface{}) []byte {
	data, err := json.Marshal(v)
	PanicIf(err)
	return data
}

func UnmarshalJson(data []byte, v interface{}) {
	err := json.Unmarshal(data, v)
	PanicIf(err)
}
