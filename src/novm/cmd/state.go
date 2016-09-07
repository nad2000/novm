package cmd

import (
	"io/ioutil"
	"github.com/rocksolidlabs/novm/src/novm/utils"
	"syscall"

	"github.com/rocksolidlabs/novm/src/novmm/control"
	"github.com/rocksolidlabs/novm/src/novmm/machine"
	"github.com/rocksolidlabs/novm/src/novmm/platform"
)

func CreateUserMemory(size int, fd int) {

}

func CreateState(
	com1 bool, com2 bool, nopci bool, memsize int) control.State {

	devices := []machine.DeviceInfo{
		{
			Name:   utils.Uuid(),
			Driver: "bios",
			Data:   map[string]interface{}{},
			Debug:  false,
		},
		{
			Name:   utils.Uuid(),
			Driver: "acpi",
			Data:   map[string]interface{}{},
			Debug:  false,
		},
		{
			Name:   utils.Uuid(),
			Driver: "apic",
			Data:   map[string]interface{}{},
			Debug:  false,
		},
		{
			Name:   utils.Uuid(),
			Driver: "pit",
			Data:   map[string]interface{}{},
			Debug:  false,
		},
	}

	if com1 {
		devices = append(devices, machine.DeviceInfo{
			Name:   utils.Uuid(),
			Driver: "uart",
			Data: map[string]interface{}{
				"base":      0x3f8,
				"interrupt": 4,
			},
			Debug: false,
		})
	}

	if com2 {
		devices = append(devices, machine.DeviceInfo{
			Name:   utils.Uuid(),
			Driver: "uart",
			Data: map[string]interface{}{
				"base":      0x2f8,
				"interrupt": 3,
			},
			Debug: false,
		})
	}

	devices = append(devices, machine.DeviceInfo{
		Name:   utils.Uuid(),
		Driver: "rtc",
		Data:   map[string]interface{}{},
		Debug:  false,
	})

	if !nopci {
		devices = append(devices, machine.DeviceInfo{
			Name:   utils.Uuid(),
			Driver: "pci-bus",
			Data:   map[string]interface{}{},
			Debug:  false,
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

	devices = append(devices, machine.DeviceInfo{
			Name:   utils.Uuid(),
			Driver: "user-memory",
			Data: map[string]interface{}{
				"fd":   fd,
				"size": 1024 * 1024 * memsize,
			},
			Debug: false,
		},
		machine.DeviceInfo{
			Name:   utils.Uuid(),
			Driver: "virtio-pci-console",
			Data:   map[string]interface{}{},
			Debug:  false,
		},
		{
			Name:   utils.Uuid(),
			Driver: "virtio-pci-fs",
			Data: map[string]interface{}{
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
			Debug: false,
		},
		{
			Name: utils.Uuid(),

			Driver: "virtio-pci-fs",
			Data: map[string]interface{}{
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
			Debug: false,
		},
	}...)

	state := control.State{
		Devices: devices,
		Vcpus:   []platform.VcpuInfo{{}},
	}
	return state
}
