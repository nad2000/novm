package cmd

import (
	"io/ioutil"
	"syscall"

	"github.com/rocksolidlabs/novm/src/novm/devices"
	"github.com/rocksolidlabs/novm/src/novm/utils"
)

func CreateUserMemory(size int, fd int) {

}

func CreateState(
	com1 bool, com2 bool, nopci bool, memsize int) devices.State {

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
		devs = append(devs, devices.State{
			"name":   utils.Uuid(),
			"driver": "uart",
			"data": map[string]interface{}{
				"base":      0x3f8,
				"interrupt": 4,
			},
			"debug": false,
		})
	}

	if com2 {
		devs = append(devs, devices.State{
			"name":   utils.Uuid(),
			"driver": "uart",
			"data": map[string]interface{}{
				"base":      0x2f8,
				"interrupt": 3,
			},
			"debug": false,
		})
	}

	devs = append(devs, devices.State{
		"name":   utils.Uuid(),
		"driver": "rtc",
		"data":   map[string]interface{}{},
		"debug":  false,
	})

	if !nopci {
		devs = append(devs, devices.State{
			"name":   utils.Uuid(),
			"driver": "pci-bus",
			"data":   map[string]interface{}{},
			"debug":  false,
		})
	}

	var fd int

	if fd < 4 {
		tf, _ := ioutil.TempFile("", "tmp")
		fd, _ = syscall.Dup(int(tf.Fd()))
		utils.ClearCloexec(fd)
	}

	if memsize < 128 {
		memsize = 1024
	}

	devs = append(devs, []devices.State{
		{
			"name":   utils.Uuid(),
			"driver": "user-memory",
			"data": map[string]interface{}{
				"fd":   fd,
				"size": 1024 * 1024 * memsize,
			},
			"debug": false,
		},
		{
			"name":   utils.Uuid(),
			"driver": "virtio-pci-console",
			"data":   map[string]interface{}{},
			"debug":  false,
		},
		{
			"name":   utils.Uuid(),
			"driver": "virtio-pci-fs",
			"data": map[string]interface{}{
				"fdlimit": "0", "read": map[string]interface{}{
					"/lib/modules/3.19.0-32-generic": []interface{}{
						"/home/nad2000/.novm/kernels/172c0b9d54445b91da234acca4f3beb31240f795/modules",
					},
					"/": []interface{}{
						"/",
					},
				},
				"tag": "root", "write": map[string]interface{}{
					"/": "/home/nad2000/.novm/instances/14185",
				},
			},
			"debug": false,
		},
		{
			"name": utils.Uuid(),

			"driver": "virtio-pci-fs",
			"data": map[string]interface{}{
				"fdlimit": "0",
				"read": map[string]interface{}{
					"/": []interface{}{},
					"/init": []interface{}{
						"/home/nad2000/novm/bin/noguest",
					},
				},
				"tag": "init", "write": map[string]interface{}{
					"/": "/tmp/tmpTi8NDV",
				},
			},
			"debug": false,
		},
	}...)

	state := devices.State{
		"devices": devs,
		"vcpus":   []interface{}{}, // TODO: number of vcpus
	}
	return state
}
