package devices

import (
	"crypto/rand"
	"fmt"
)

// Ioctl numbers are encoded in 32 bits:
//   1  - 16: command
//   17 - 30: param struct size
//   31 - 32: direction
func iow(cmdType int, nr int, size int) uint32 {
	return uint32((1 << 30) | (size << 16) | nr | (cmdType << 8))
}
func ior(cmdType int, nr int, size int) uint32 {
	return uint32((2 << 30) | (size << 16) | nr | (cmdType << 8))
}

const (
	//  Tap device flags:
	IFF_TAP      = 0x0002
	IFF_NO_PI    = 0x1000
	IFF_VNET_HDR = 0x4000

	// Tap device offloads:
	TUN_F_CSUM    = 0x01
	TUN_F_TSO4    = 0x02
	TUN_F_TSO6    = 0x04
	TUN_F_TSO_ECN = 0x08
	TUN_F_UFO     = 0x10

	// Tap device ioctls:
	TUNSETIFF       = iow('T', 202, 4)
	TUNGETFEATURES  = ior('T', 207, 4)
	TUNSETOFFLOAD   = iow('T', 208, 4)
	TUNGETVNETHDRSZ = ior('T', 215, 4)
)

// Return a random MAC address
func randomMac(oui string) {
	if oui == "" {
		oui = "28:48:46"
	}

	b := make([]byte, 3)
	rand.Read(b)
	return fmt.Sprintf("%s:%02x:%02x:%02x", oui, b[0], b[1], b[2])
}

// Parse an IP address given CIDR form.
//
// We also return the associated gateway
// and broadcast address for the subnet.
func parseIpv4mask(ip string) (string, string, string) {
	return "", "", ""
}
