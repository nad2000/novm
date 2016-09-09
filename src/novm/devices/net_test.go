package devices_test

import (
	"fmt"
	"testing"

	"github.com/rocksolidlabs/novm/src/novm/devices"
)

func ExampleParseIpv4mask() {
	fmt.Println(devices.ParseIpv4mask("10.0.0.1/16"))
	fmt.Println(devices.ParseIpv4mask("10.11.22.33/18"))
	fmt.Println(devices.ParseIpv4mask("10.11.22.33/24"))
	// Output:
	// 10.0.0.1 10.0.0.1 10.0.255.254
	// 10.11.22.33 10.11.0.1 10.11.63.254
	// 10.11.22.33 10.11.22.1 10.11.22.254
}

func TestTUNCommands(t *testing.T) {
	for _, r := range []struct {
		name     string
		value    int
		expected int
	}{
		{"TUNSETIFF", devices.TUNSETIFF, 1074025674},
		{"TUNGETFEATURES", devices.TUNGETFEATURES, 2147767503},
		{"TUNSETOFFLOAD", devices.TUNSETOFFLOAD, 1074025680},
		{"TUNGETVNETHDRSZ", devices.TUNGETVNETHDRSZ, 2147767511},
	} {
		if r.value != r.expected {
			t.Fatalf("Wrong value for %#v: %d, want: %d", r.name, r.value, r.expected)
		}
	}
}
