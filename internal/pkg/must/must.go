package must

import (
	"encoding/json"
	"gopkg.in/yaml.v3"
)

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

