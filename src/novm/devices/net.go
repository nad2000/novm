//golint:ignore
package devices

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"unsafe"

	log "github.com/Sirupsen/logrus"
	"github.com/kopwei/goovs"
	"github.com/vishvananda/netlink"

	"github.com/rocksolidlabs/novm/src/novm/utils"
)

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

// Return a random MAC address
func randomMac(oui string) string {
	if oui == "" {
		oui = "28:48:46"
	}

	b := make([]byte, 3)
	rand.Read(b)
	return fmt.Sprintf("%s:%02x:%02x:%02x", oui, b[0], b[1], b[2])
}

// ItoIpv4 converts 32 bit integer into string representation of IPv4 address
func ItoIpv4(addr int) string {
	return strings.Join([]string{
		strconv.Itoa((addr >> 24) & 0xff),
		strconv.Itoa((addr >> 16) & 0xff),
		strconv.Itoa((addr >> 8) & 0xff),
		strconv.Itoa(addr & 0xff),
	}, ".")
}

// ParseIpv4mask parses an IP address given CIDR form.
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

// TapDevice creates a TAP device
func TapDevice(name string) (tap *os.File, vnet int, offload bool) {

	tap, err := os.OpenFile("/dev/net/tun", os.O_RDWR|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("Failed to open '/dev/net/tun':", err)
	}

	// Figure out if the kernel supports processing vnet headers on tap
	// devices. This is necessary for forwarding hardware offloading
	// from the guest virtual nics to the physical nics on the host.
	var features int
	r, _, e1 := syscall.Syscall(
		syscall.SYS_IOCTL,
		uintptr(tap.Fd()),
		syscall.TUNGETFEATURES,
		uintptr(unsafe.Pointer(&features)))

	if r != 0 {
		log.Fatal(e1)
	}

	flags := IFF_TAP | IFF_NO_PI
	if features&IFF_VNET_HDR != 0 {
		flags |= IFF_VNET_HDR
	}

	// Ctreate the TAP device
	var ifr [18]byte
	copy(ifr[:], []byte(name))
	binary.LittleEndian.PutUint16(ifr[16:], uint16(flags))
	r, _, e1 = syscall.Syscall(
		syscall.SYS_IOCTL,
		uintptr(tap.Fd()),
		syscall.TUNSETIFF,
		uintptr(unsafe.Pointer(&ifr)))
	if r != 0 {
		log.Fatal(e1)
	}
	var vnetHdrSz int
	r, _, e1 = syscall.Syscall(
		syscall.SYS_IOCTL,
		uintptr(tap.Fd()),
		syscall.TUNGETVNETHDRSZ,
		uintptr(unsafe.Pointer(&vnetHdrSz)))
	if r != 0 {
		log.Fatal(e1)
	}

	// Size of the vnet header expected by the TAP device:
	if features&IFF_VNET_HDR != 0 {
		vnet = vnetHdrSz
		r, _, e1 = syscall.Syscall(
			syscall.SYS_IOCTL,
			uintptr(tap.Fd()),
			syscall.TUNSETOFFLOAD,
			uintptr(TUN_F_CSUM|TUN_F_TSO4|TUN_F_TSO6|TUN_F_TSO_ECN|TUN_F_UFO))
		if offload = r == 0; !offload {
			log.Print("Failed to enable offloads:", e1)
		}
	} // else vnet=0, offload=false

	return tap, vnet, offload
}

// CreateNic creates a vitio network device.
func CreateNic(index int, options map[string]string) (state State, ip string) {
	var (
		mac, tapName, bridge, gateway, mtu, name, driver string
		ok, debug                                        bool
	)

	_ = gateway

	debug = utils.Asbool(options["debug"])

	name, ok = options["name"]
	if !ok || name == "" {
		name = utils.UUID()
	}

	mac = options["mac"]
	if mac == "" {
		mac = randomMac("")
	}
	tapName = options["tapname"]
	if tapName == "" {
		tapName = fmt.Sprintf("novm%d-%d", os.Getpid(), index)
	}
	tap, vnet, offload := TapDevice(tapName)
	mtu, ok = options["mtu"]
	if ok && mtu != "" {
		command := exec.Command("/sbin/ip", "link", "set", "dev", tapName, "mtu", mtu)
		out, err := command.Output()
		if err != nil {
			log.Fatal("Failed to set up MTU:", err, "/", out)
		}
	}

	// Enslave to the given bridge.
	bridge, ok = options["bridge"]
	if ok && bridge != "" {
		driver, ok = options["driver"]
		if !ok && driver == "" {
			driver = "bridge"
		}
		err := EnslaveNic(driver, tapName, bridge)
		if err != nil {
			log.Fatalf("Failed to add interface %#v to bridge %#v: %v", tapName, bridge, err)
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
	ip = options["ip"]
	if ip != "" {
		address, start, end := ParseIpv4mask(ip)
		dnsmasqArgs := []string{
			"--keep-in-foreground",
			"--no-daemon",
			"--conf-file=",
			"--bind-interfaces",
			"--except-interface=lo",
			"--interface=" + tapName,
			"--dhcp-range=" + start + "," + end,
			"--dhcp-host=" + mac + "," + address + "," + name,
		}
		if gateway != "" {
			dnsmasqArgs = append(dnsmasqArgs, "--dhcp-option=option:router,"+gateway)
		}

		command := exec.Command("/usr/sbin/dnsmasq", dnsmasqArgs...)
		command.Stdout = os.Stdout
		command.Stderr = os.Stderr
		command.Start()
	}

	fd := tap.Fd()
	utils.ClearCloexec(int(fd))

	state = createVirtio(index,
		map[string]interface{}{
			"fd":      fd,
			"mac":     mac,
			"vnet":    vnet,
			"offload": offload,
			"ip":      ip,
		}, "net", name)
	state["debug"] = debug

	return
}

// EnslaveNic enslaves a nic to a supplied bridge.
func EnslaveNic(driver, tapName, bridge string) error {
	if driver == "bridge" {
		// check if the bridge exists
		_, err := netlink.LinkByName(bridge)
		// create the bridge
		if err != nil {
			la := netlink.NewLinkAttrs()
			la.Name = bridge
			br := &netlink.Bridge{la}
			err := netlink.LinkAdd(br)
			if err != nil {
				return fmt.Errorf("Error creating Linux Bridge: %s", err)
			}
		}
		// enslave the tuntap
		_, err = netlink.LinkByName(bridge)
		if err != nil {
			return fmt.Errorf("Error getting linux bridge: %s, %s", bridge, err)
		}
		tap, err := netlink.LinkByName(tapName)
		if err != nil {
			return fmt.Errorf("Error getting linux tuntap: %s, %s", tapName, err)
		}
		err = netlink.LinkSetMaster(tap, &netlink.Bridge{netlink.LinkAttrs{Name: bridge}})
		if err != nil {
			return fmt.Errorf("Error enslaving tuntap to Linux bridge: %s -> %s, %s", tapName, bridge, err)
		}
	} else if driver == "ovs" {
		client, err := goovs.GetOVSClient("unix", "")
		if err != nil {
			return fmt.Errorf("Error obtaining unix socket connection to Open vSwitch DB: %s", err)
		}
		// check if bridge exists
		ok, err := client.BridgeExists(bridge)
		if err != nil {
			return fmt.Errorf("Error checking if Open vSwitch bridge exists: %s", err)
		}
		// create the bridge
		if !ok {
			err := client.CreateBridge(bridge)
			if err != nil {
				return fmt.Errorf("Error creating Open vSwitch bridge: %s", err)
			}
		}
		// enslave the tuntap
		err = client.CreateVethPort(bridge, tapName, 0)
		if err != nil {
			return fmt.Errorf("Error enslaving tuntap to Open vSwitch bridge: %s -> %s, %s", tapName, bridge, err)
		}
	} else {
		return fmt.Errorf("Please supply a valid networking driver: bridge|ovs")
	}

	return nil
}
