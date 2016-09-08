package devices

import (
	"crypto/rand"
	"fmt"
	"strconv"
	"strings"
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
)

var (
	// Tap device ioctls:
	TUNSETIFF       = iow('T', 202, 4)
	TUNGETFEATURES  = ior('T', 207, 4)
	TUNSETOFFLOAD   = iow('T', 208, 4)
	TUNGETVNETHDRSZ = ior('T', 215, 4)
)

// Return a random MAC address
func randomMac(oui string) string {
	if oui == "" {
		oui = "28:48:46"
	}

	b := make([]byte, 3)
	rand.Read(b)
	return fmt.Sprintf("%s:%02x:%02x:%02x", oui, b[0], b[1], b[2])
}

func ItoIpv4(addr int) string {
	return strings.Join([]string{
		strconv.Itoa((addr >> 24) & 0xff),
		strconv.Itoa((addr >> 16) & 0xff),
		strconv.Itoa((addr >> 8) & 0xff),
		strconv.Itoa(addr & 0xff),
	}, ".")
}

// Parse an IP address given CIDR form.
//
// We also return the associated gateway
// and broadcast address for the subnet.
func ParseIpv4mask(ip string) (string, string, string) {
	// Compute the relevant masks
	parts := strings.Split(ip, "/")
	addrPart, maskPart := parts[0], parts[1]
	addr := 0
	for i, p := range strings.Split(addrPart, ".") {
		pInt, _ := strconv.Atoi(p)
		addr += pInt << uint(24-i*8)
	}
	maskInt, _ := strconv.Atoi(maskPart)
	mask := ((1 << uint(maskInt)) - 1) << uint(32-maskInt)

	// Compute the addresses
	networkAddr := addr & mask
	firstAddr := networkAddr + 1
	broadcastAddr := networkAddr + (^mask & (mask - 1))
	endAddr := broadcastAddr - 1

	return addrPart, ItoIpv4(firstAddr), ItoIpv4(endAddr)
}
