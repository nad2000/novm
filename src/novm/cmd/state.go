package cmd

import (
	"os"
	"strconv"

	"github.com/rocksolidlabs/novm/src/novm/devices"
	"github.com/rocksolidlabs/novm/src/novm/utils"
)

func CreateState(
	cpus int, com1 bool, com2 bool, nopci bool, memsize int,
	nics []string, disks []string) devices.State {

	if cpus <= 0 {
		cpus = 1
	}

	devices.Pci = !nopci

	devs := []devices.State{
		{
			"name":   utils.Uuid(),
			"driver": "bios",
			"data":   map[string]interface{}{},
			"debug":  false,
		},
		{
			"name":   utils.Uuid(),
			"driver": "acpi",
			"data":   map[string]interface{}{},
			"debug":  false,
		},
		{
			"name":   utils.Uuid(),
			"driver": "apic",
			"data":   map[string]interface{}{},
			"debug":  false,
		},
		{
			"name":   utils.Uuid(),
			"driver": "pit",
			"data":   map[string]interface{}{},
			"debug":  false,
		},
	}

	if com1 {
		devs = append(devs, devices.CreateCom1())
	}

	if com2 {
		devs = append(devs, devices.CreateCom2())
	}

	devs = append(devs, devices.State{
		"name":   utils.Uuid(),
		"driver": "rtc",
		"data":   map[string]interface{}{},
		"debug":  false,
	})

	if devices.Pci {
		devs = append(devs, devices.State{
			"name":   utils.Uuid(),
			"driver": "pci-bus",
			"data":   map[string]interface{}{},
			"debug":  false,
		})
	}

	var fd int
	if memsize < 128 {
		memsize = 1024
	}

	devs = append(devs, []devices.State{
		devices.CreateUserMemory(1024*1024*memsize, fd),
		// Always enable the console.
		// The noguest binary that executes inside
		// the guest will use this as an RPC mechanism.
		devices.CreateConsole(0),
	}...)

	// TODO:
	// Build our NICs

	// TODO:
	// Build our disks:

	// TODO:
	// Add modules:

	// TODO:
	// Add our packs.
	// NOTE: All packs are given relative to root.

	// TODO:
	// Add our docker repositories
	// NOTE: These are also all relative to root.

	// TODO (add all packages, ...):
	read := []string{}
	write := []string{}
	// The root filesystem:
	rootFs := devices.CreateFs(1+len(nics)+len(disks),
		"root", instances.File(strconv.Itoa(os.Getpid())), read, write)
	devs = append(devs, rootFs)

	// Add noguest as our init:
	initFs := devices.CreateFs(2+len(nics)+len(disks),
		"init", "", []string{"/init=>" + utils.LibExec("noguest")}, nil)
	devs = append(devs, initFs)
	vcpus := make([]devices.State, cpus)
	for i := 0; i < cpus; i++ {
		vcpus[i] = devices.State{}
	}
	state := devices.State{
		"devices": devs,
		"vcpus":   vcpus,
	}
	return state
}
