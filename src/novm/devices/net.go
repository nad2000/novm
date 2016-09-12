package devices

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"unsafe"
)

// Ioctl numbers are encoded in 32 bits:
//   1  - 16: command
//   17 - 30: param struct size
//   31 - 32: direction
func iow(cmdType int, nr int, size int) int {
	return ((1 << 30) | (size << 16) | nr | (cmdType << 8))
}
func ior(cmdType int, nr int, size int) int {
	return ((2 << 30) | (size << 16) | nr | (cmdType << 8))
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

func fcntl(fd uintptr, cmd int, arg uintptr) (val int, err error) {
	r, _, e := syscall.Syscall(syscall.SYS_FCNTL, uintptr(fd), uintptr(cmd), arg)
	val = int(r)
	if e != 0 {
		switch e {
		case 0xb:
			err = syscall.EAGAIN
		case 0x16:
			err = syscall.EINVAL
		case 0x2:
			err = syscall.ENOENT
		default:
			err = error(e)
		}
	}
	return
}

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

// TapDevice creats a TAP device
func TapDevice(name string) (tap *os.File, vnet int, offload bool) {

	tap, err := os.OpenFile("/dev/net/tun", os.O_RDWR, 0666)
	if err != nil {
		log.Fatal("Failed to open '/dev/net/tun':", err)
	}

	// Figure out if the kernel supports processing vnet headers on tap
	// devices. This is necessary for forwarding hardware offloading
	// from the guest virtual nics to the physical nics on the host.
	features, err := fcntl(tap.Fd(), TUNGETFEATURES, uintptr(0))

	flags := IFF_TAP | IFF_NO_PI | (features & IFF_VNET_HDR)

	// Ctreate the TAP device
	var ifr [18]byte
	copy(ifr[:], []byte(name))
	binary.LittleEndian.PutUint16(ifr[:], uint16(flags))
	fcntl(tap.Fd(), TUNSETIFF, uintptr(unsafe.Pointer(&ifr)))
	vnetHdrSz, _ := fcntl(tap.Fd(), TUNGETVNETHDRSZ, uintptr(0))

	// Size of the vnet header expected by the TAP device:
	if features&IFF_VNET_HDR != 0 {
		vnet = vnetHdrSz
		_, err := fcntl(tap.Fd(), TUNSETOFFLOAD,
			uintptr(TUN_F_CSUM|TUN_F_TSO4|
				TUN_F_TSO6|TUN_F_TSO_ECN|TUN_F_UFO))
		if offload = err == nil; !offload {
			log.Println("Failed to enable offloads:", err)
		}
	} // else vnet=0, offload=false

	return tap, vnet, offload
}

// CreateNic creates a vitio network device.
func CreateNic(index int, mac string, tapName string, bridge string, ip string, gateway string, mtu int,
	name string) (state State) {
	if mac == "" {
		mac = randomMac("")
	}
	if tapName == "" {
		tapName = fmt.Sprintf("novm%d-%d", os.Getpid(), index)
	}
	tap, vnet, offload := TapDevice(tapName)
	if mtu != 0 {
		command := exec.Command("/sbin/ip", "link", "set", "dev", tapName, "mtu", strconv.Itoa(mtu))
		out, err := command.Output()
		if err != nil {
			log.Fatal("Failed to set up MTU:", err, "/", out)
		}
	}

	// Enslave to the given bridge. (It will be authomatically removed.)
	if bridge != "" {
		command := exec.Command("/sbin/brctl", "addif", bridge, tapName)
		out, err := command.Output()
		if err != nil {
			log.Fatalf("Failed to add interface %#v to bridge %#v: %v (%v)", tapName, bridge, err, out)
		}
	}

	// Make sure the interface is up:
	command := exec.Command("/sbin/ip", "link", "set", "up", "dev", tapName)
	out, err := command.Output()
	if err != nil {
		log.Fatalf("Failed to set up the interface %#v: %v (%v) ", tapName, err, out)
	} else {
		log.Println(out)
	}
	// Start our dnsmasq.
	// This is just a simple responded for
	// DHCP queries and routes that will die
	// whenever the underlying instance dies.
	if ip != "" {
		// TODO:
	}

	var driver string
	if Pci { //
		driver = "virtio-pci-net"
	} else {
		driver = "virtio-mmio-net"
	}
	fd := tap.Fd()

	return State{
		"cmdline": nil, // TODO!
		"data": map[string]interface{}{
			"fd":      fd,
			"mac":     mac,
			"vnet":    vnet,
			"offload": offload,
			"ip":      ip,
		},
		"debug": false,
		"drive": driver,
		"name":  name,
	}

}
