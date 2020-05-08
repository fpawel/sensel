package pkg

import (
	"fmt"
	"github.com/fpawel/comm"
	"io"
	"time"
)

func MockComm(f func([]byte) []byte) comm.T {
	return comm.New(MockComport(f), comm.Config{
		TimeoutGetResponse: time.Millisecond,
		TimeoutEndResponse: 0,
	})
}

func MockComport(f reqToRespFunc) io.ReadWriter {
	return &mockComport{f: f}
}

type reqToRespFunc = func(req []byte) []byte

type mockComport struct {
	req  []byte
	resp []byte
	f    reqToRespFunc
}

func (x *mockComport) Write(p []byte) (int, error) {
	x.req = p
	x.resp = x.f(x.req)
	return len(p), nil
}

func (x *mockComport) Read(p []byte) (int, error) {
	if len(x.resp) == 0 {
		return 0, fmt.Errorf("unsupported request %02X", x.req)
	}
	if len(p) < len(x.resp) {
		return len(x.resp), nil
	}
	return copy(p, x.resp), nil
}
