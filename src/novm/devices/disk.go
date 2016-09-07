package devices

import (
	"novm/utils"
	"os"
	"syscall"
)

// CreateDisk craates a disk and returns its state:
func CreateDisk(index int, fileName string, dev string, name string, pci bool) State {

	if fileName == "" {
		fileName = "/dev/null"
	}

	if dev == "" {
		dev = "vd" + string('a'+index)
	}

	// Open the device:
	f := os.Open(fileName)
	fd, _ = syscall.Dup(int(f.Fd()))
	utils.ClearCloexec(fd)

	driver := "virtio-"
	if pci {
		driver += "pci"
	} else {
		driver += "mmio"
	}
	dirver += "-block"

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
