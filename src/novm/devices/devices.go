package devices

import (
	"fmt"
	"io/ioutil"
	"log"
	"syscall"

	"github.com/rocksolidlabs/novm/src/novm/utils"
)

var (
	// Debug global debugging option
	Debug bool
	// Pci PCI flag. If 'true' PIC drivers would be set, if 'false' - MMIO
	Pci bool
)

// State is a shorthand equivalent for Python dictionary used as "state"
type State map[string]interface{}

func createDevice(driver string, name string, data State, cmdline string) State {

	if data == nil {
		data = State{}
	}

	if name == "" {
		name = utils.UUID()
	}

	if driver == "" {
		log.Fatal("Missing driver for NOVMM.")
	}

	device := State{
		"driver": driver,
		"name":   name,
		"data":   data,
		"debug":  Debug,
	}
	if cmdline != "" {
		device["cmdline"] = cmdline
	}

	return device
}

func createVirtio(index int, data State, virtioDriver string, name string) (device State) {

	var (
		driver  string
		cmdline string
	)

	if data == nil {
		data = State{}
	}

	if Pci {
		driver = "virtio-pci-" + virtioDriver
	} else {
		driver = "virtio-mmio-" + virtioDriver
	}

	if index >= 0 && !Pci {
		data["address"] = 0xe0000000 + index*4096
		data["interrupt"] = 32 + index
		cmdline = fmt.Sprintf("virtio-mmio.%d@0x%x:%d:%d",
			index, data["address"], data["interrupt"], index)
	} else {
		cmdline = ""
	}

	return createDevice(driver, name, data, cmdline)
}

func createCom(comNo int) State {
	var (
		base uint16 // I/O port
		irq  uint16 // interrupt
	)
	switch comNo {
	case 1:
		base, irq = 0x3f8, 4
	case 2:
		base, irq = 0x2f8, 3
	case 3:
		base, irq = 0x3e8, 4
	case 4:
		base, irq = 0x2e8, 3
	}

	return State{
		"driver": "uart",
		"name":   utils.UUID(),
		"data": State{
			"base":      base,
			"interrupt": irq,
		},
		"cmdline": fmt.Sprintf("console=uart,io,0x%03x", base),
		"debug":   Debug,
	}
}

// CreateCom1 creats serial port COM1
func CreateCom1() State {
	return createCom(1)
}

// CreateCom2 creats serial port COM2
func CreateCom2() State {
	return createCom(2)
}

// CreateConsole creates a virtio serial/console device:
func CreateConsole(index int) State {
	return createVirtio(index, nil, "console", "")
}

// CreateUserMemory create user memory (with size in MiBs)
func CreateUserMemory(size int, fd int) State {

	if fd < 4 {
		tf, _ := ioutil.TempFile("", "tmp")
		fd, _ = syscall.Dup(int(tf.Fd()))
		utils.ClearCloexec(fd)
	}

	return State{
		"name":   utils.UUID(),
		"driver": "user-memory",
		"data": map[string]interface{}{
			"fd":   fd,
			"size": size,
		},
		"debug": Debug,
	}
}
