package cmd

import (
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	log "github.com/Sirupsen/logrus"

	"github.com/rocksolidlabs/novm/src/novm/devices"
	"github.com/rocksolidlabs/novm/src/novm/docker"
	"github.com/rocksolidlabs/novm/src/novm/utils"
)

// CreateInstance creates VM and stores the configuration in the state object
func CreateInstance(
	output io.Writer,
	kernel string, release string, cpus int, com1 bool, com2 bool, nopci bool, memsize int,
	nics []string, disks []string, read []string, write []string, packList []string, packageList []string,
	repos []string, name string, cmdline string) (instanceID string, devCmdline string, meta devices.State) {

	instanceID = utils.UUID()

	if cpus <= 0 {
		cpus = 1
	}

	if read == nil {
		read = make([]string, 0, 10)
	}

	if write == nil {
		write = make([]string, 0, 10)
	}

	if cmdline != "" && !strings.HasPrefix(cmdline, " ") {
		cmdline = " " + cmdline
	}

	devices.Pci = !nopci

	devs := []devices.State{
		{
			"name":    utils.UUID(),
			"driver":  "bios",
			"data":    map[string]interface{}{},
			"cmdline": nil,
			"debug":   false,
		},
		{
			"name":    utils.UUID(),
			"driver":  "acpi",
			"data":    map[string]interface{}{},
			"cmdline": "intel_pstate=disable",
			"debug":   false,
		},
		{
			"name":    utils.UUID(),
			"driver":  "apic",
			"data":    map[string]interface{}{},
			"cmdline": nil,
			"debug":   false,
		},
		{
			"name":    utils.UUID(),
			"driver":  "pit",
			"data":    map[string]interface{}{},
			"cmdline": nil,
			"debug":   false,
		},
	}

	if com1 {
		devs = append(devs, devices.CreateCom1())
	}

	if com2 {
		devs = append(devs, devices.CreateCom2())
	}

	devs = append(devs, devices.State{
		"name":    utils.UUID(),
		"driver":  "rtc",
		"data":    map[string]interface{}{},
		"cmdline": nil,
		"debug":   false,
	})

	if devices.Pci {
		devs = append(devs, devices.State{
			"name":    utils.UUID(),
			"driver":  "pci-bus",
			"cmdline": "pci=conf1",
			"data":    map[string]interface{}{},
			"debug":   false,
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

	// Build our NICs
	ips := []string{}
	for index, nic := range nics {
		options := make(map[string]string)
		for _, opt := range strings.Split(nic, ",") {
			parts := strings.SplitN(opt, "=", 2)
			if len(parts) < 2 {
				log.Fatalf("Malformated NIC option: %q", opt)
			}
			options[parts[0]] = parts[1]
		}
		nic, ip := devices.CreateNic(index, options)
		devs = append(devs, nic)
		if ip != "" {
			ips = append(ips, ip)
		}
	}

	// Build our disks:
	for index, disk := range disks {
		options := make(map[string]string)
		for _, opt := range strings.Split(disk, ",") {
			parts := strings.SplitN(opt, "=", 2)
			if len(parts) < 2 {
				log.Fatalf("Malformated disk option: %q", opt)
			}
			options[parts[0]] = parts[1]
		}
		devs = append(devs, devices.CreateDisk(index+len(nics)+1, options))
	}

	// Add modules:
	if moduleDir := kernels.File(kernel, "modules"); utils.PathExists(moduleDir) {
		read = append(read, fmt.Sprintf("/lib/modules/%s=>%s", release, moduleDir))
	}

	// Add our packs.
	// NOTE: All packs are given relative to root.
	for _, p := range packList {
		read = append(read, packs.File(p))
	}

	// Add our docker repositories
	// NOTE: These are also all relative to root.
	for _, d := range repos {
		opts := strings.Split(d, ",")
		repository := opts[0]
		options := make(map[string]string)
		for _, arg := range opts[1:] {
			parts := strings.SplitN(arg, "=", 2)
			options[parts[0]] = parts[1]
		}
		docker.Init(dockerDB, options)
		read = append(read, docker.PullRepository(repository)...)
	}

	// The root filesystem:
	rootFs := devices.CreateFs(1+len(nics)+len(disks),
		"root", instances.File(instanceID), read, write)
	devs = append(devs, rootFs)

	// Add noguest as our init:
	initFs := devices.CreateFs(2+len(nics)+len(disks),
		"init", "", []string{"/init=>" + utils.LibExec("noguest")}, nil)
	devs = append(devs, initFs)

	// Add packages:
	log.Debug(packageList)
	for idx, id := range packageList {
		p, ok := packages.Get(id)
		if !ok {
			log.Errorf("Failed to find package %q", id)
			continue
		}
		mountPoint := p["mount"]
		name := p["name"]
		root := path.Join(packages.File(id), "ROOT")

		log.Debugf("Adding package %q (%q) mounting on %q", name, root, mountPoint)

		packageFs := devices.CreateFs(idx+3+len(nics)+len(disks), id,
			root, []string{mountPoint}, nil)
		devs = append(devs, packageFs)
	}
	vcpus := make([]devices.State, cpus)
	for i := 0; i < cpus; i++ {
		vcpus[i] = devices.State{}
	}
	state := devices.State{
		"devices": devs,
		"vcpus":   vcpus,
	}
	utils.JSONDump(state, output, true)

	meta = map[string]interface{}{
		"name":   name,
		"cpus":   cpus,
		"memory": memsize,
		"kernel": kernel,
		"ips":    ips,
		"pid":    os.Getpid(),
	}

	devCmdline = "-cmdline="
	for _, d := range devs {
		cl := d.Cmdline()
		if cl != "" {
			if !strings.HasSuffix(devCmdline, "=") {
				devCmdline += " "
			}
			devCmdline += cl
		}
	}
	devCmdline += cmdline
	return

}
