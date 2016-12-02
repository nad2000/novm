package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"syscall"

	log "github.com/Sirupsen/logrus"

	"github.com/rocksolidlabs/novm/src/novm/devices"
	"github.com/rocksolidlabs/novm/src/novm/docker"
	"github.com/rocksolidlabs/novm/src/novm/utils"
)

type InstaceParams struct {
	Kernel, Release, Name, Cmdline   string
	Cpus, Memsize                    int
	Com1, Com2, Nopci                bool
	Nics, Disks, Read, Write, Images []string
	Repos                            []string
	Packs                            []string
	Packages                         []string
}

// CreateInstance creates VM and stores the configuration in the state object
func CreateInstance(params InstaceParams) (stateFd int, instanceID string, devCmdline string) {
	instanceID = utils.NewID()
	stateFd, devCmdline = ActivateInstance(instanceID, params)
	return
}

// ActivateInstance activates an instace with given ID VM
// and stores the configuration in the state object
// it attempts to reuse resouces linked to the give ID
func ActivateInstance(instanceID string, params InstaceParams) (stateFd int, devCmdline string) {

	kernel, release, cpus, com1, com2 := params.Kernel, params.Release, params.Cpus, params.Com1, params.Com2
	nics, disks, read, write, packList, packageList := params.Nics, params.Disks, params.Read, params.Write, params.Packs, params.Packages
	repos, name, cmdline, nopci, memsize := params.Repos, params.Name, params.Cmdline, params.Nopci, params.Memsize

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

	// Add our docker images
	// NOTE: These are also all relative to root.
	for _, image := range images {
		docker.Init(dockerDB, image)
		read = append(read, docker.PullImage())
	}

	// Add packages to the read map:
	for _, id := range packageList {
		p, ok := packages.Get(id)
		if !ok {
			log.Errorf("Failed to find package %q", id)
			continue
		}
		mountPoint := p["mount"]
		name := p["name"]
		root := path.Join(packages.File(id), "ROOT")

		log.Debugf("Adding package %q (%q) mounting on %q", name, root, mountPoint)

		read = append(read, mountPoint+"=>"+root)
	}

	// The root filesystem:
	rootFs := devices.CreateFs(1+len(nics)+len(disks),
		"root", instances.File(instanceID), read, write)
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

	// Provide our state by dumping to a temporary file:
	sf, _ := ioutil.TempFile("", "state")

	if debug {
		log.Info("Sate stored in: ", sf.Name())
	} else {
		os.Remove(sf.Name())
	}

	utils.JSONDump(state, sf, true)
	sf.Seek(0, 0)

	stateFd, _ = syscall.Dup(int(sf.Fd()))
	utils.ClearCloexec(stateFd)
	sf.Close()

	paramFileName := instances.File("." + instanceID + ".params")
	pf, _ := os.Create(paramFileName)
	defer pf.Close()
	json.NewEncoder(pf).Encode(params)

	instances.Add(instanceID, map[string]interface{}{
		"name":   name,
		"cpus":   cpus,
		"memory": memsize,
		"kernel": kernel,
		"ips":    ips,
		"pid":    os.Getpid(),
	})

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
