package main

import (
	"fmt"
	"testing"
	//   "time"
	//   "reflect"
)

func TestTCPSimple(t *testing.T) {
	rafts:=makeRafts();
	rafts=rafts
}

func expect(t *testing.T, a string, b string) {
	if a != b {
		t.Error(fmt.Sprintf("Expected %v, found %v", b, a)) // t.Error is visible when running `go test -verbose`
	}
}
