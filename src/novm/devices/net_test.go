package devices_test

import (
	"fmt"

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
