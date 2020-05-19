package winapi

import (
	"fmt"
	"testing"
)

func Test_enumPrinters(t *testing.T) {
	fmt.Println(EnumPrinters())
	fmt.Println(GetDefaultPrinter())
}
