package devices

import (
	"os"
	"syscall"

	"github.com/rocksolidlabs/novm/src/novm/utils"
)

// CreateDisk craates a disk and returns its state:
func CreateDisk(index int, options map[string]string) State {

	var (
		fileName, dev, name string
		ok, debug           bool
	)

	fileName, ok = options["filename"]
	if !ok || fileName == "" {
		fileName = "/dev/null"
	}

	dev, ok = options["dev"]
	if !ok || dev == "" {
		dev = "vd" + string('a'+index)
	}

	debug = utils.Asbool(options["debug"])

	// Open the device:
	f, _ := os.Open(fileName)
	fd, _ := syscall.Dup(int(f.Fd()))
	utils.ClearCloexec(fd)

	driver := "virtio-"
	if Pci {
		driver += "pci"
	} else {
		driver += "mmio"
	}
	driver += "-block"

	if name == "" {
		name = utils.UUID()
	}

	return State{
		"cmdline": nil,
		"data":    map[string]interface{}{"dev": dev, "fd": fd},
		"debug":   debug,
		"drive":   driver,
		"name":    name,
	}

}
