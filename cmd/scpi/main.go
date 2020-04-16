package main

import (
	"context"
	"fmt"
	"github.com/fpawel/comm"
	"github.com/fpawel/comm/comport"
	"os"
	"strings"
	"time"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("request string expected")
		return
	}
	comPort := comport.NewPort(comport.Config{
		Name:        os.Args[1],
		Baud:        115200,
		ReadTimeout: time.Millisecond,
	})
	fatalIf(comPort.Open())

	cfgComm := comm.Config{
		TimeoutEndResponse: 50 * time.Millisecond,
	}
	var err error
	cfgComm.TimeoutGetResponse, err = time.ParseDuration(os.Args[2])
	fatalIf(err)

	request := os.Args[4] + "\n"
	ctx := context.Background()

	switch os.Args[3] {
	case "W":
		err := comm.Write(ctx, []byte(request), comPort, cfgComm)
		fatalIf(err)
	case "R":
		cm := comm.New(comPort, cfgComm)
		b, err := cm.GetResponse(nil, context.Background(), []byte(request))
		fatalIf(err)
		s := string(b)
		s = strings.TrimSpace(s)
		fmt.Println(s)
	}
}

func fatalIf(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
