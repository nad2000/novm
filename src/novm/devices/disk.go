package devices

import (
	"os"
	"syscall"

	"github.com/rocksolidlabs/novm/src/novm/utils"
)

// CreateDisk craates a disk and returns its state:
func CreateDisk(index int, fileName string, dev string, name string) State {

	if fileName == "" {
		fileName = "/dev/null"
	}

	if dev == "" {
		dev = "vd" + string('a'+index)
	}

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
		name = utils.Uuid()
	}

	return State{
		"cmdline": nil,
		"data":    map[string]interface{}{"dev": dev, "fd": fd},
		"debug":   false,
		"drive":   driver,
		"name":    name,
	}

}
