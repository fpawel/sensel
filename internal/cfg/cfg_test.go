package cfg

import (
	"github.com/fpawel/sensel/internal/pkg/must"
	"testing"
)

func BenchmarkGetGob(b *testing.B) {
	for i := 0; i < b.N; i++ {
		getGob()
	}
}

func BenchmarkGetJson(b *testing.B) {
	for i := 0; i < b.N; i++ {
		getJson()
	}
}

func getJson() (r Config) {
	must.UnmarshalJson(must.MarshalJson(cfg), &r)
	return
}
