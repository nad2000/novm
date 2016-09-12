package devices_test

import (
	"testing"

	"github.com/rocksolidlabs/novm/src/novm/devices"
)

func TestCreatCom(t *testing.T) {

	com1 := devices.CreateCom1()
	if cmdline, expected := com1["cmdline"], "console=uart,io,0x3f8"; cmdline != expected {
		t.Errorf("Failed to create COM1: %#v, want: %#v (%#v)", cmdline, expected, com1)
	}
	com2 := devices.CreateCom2()
	if cmdline, expected := com2["cmdline"], "console=uart,io,0x2f8"; cmdline != expected {
		t.Errorf("Failed to create COM2: %#v, want: %#v (%#v)", cmdline, expected, com2)
	}
}
