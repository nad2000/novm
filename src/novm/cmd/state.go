package cmd

import (
	"github.com/rocksolidlabs/novm/src/novmm/control"
	"github.com/rocksolidlabs/novm/src/novmm/machine"
	"github.com/rocksolidlabs/novm/src/novmm/platform"
)

func CreateState() control.State {

	devices := []machine.DeviceInfo{
		machine.DeviceInfo{
			Name:   "36bb9417-21aa-4d80-8ffa-29290e10697a",
			Driver: "bios",
			Data:   map[string]interface{}{},
			Debug:  false,
		},
		machine.DeviceInfo{
			Name:   "ee3068ec-bfa4-43d0-b33b-4bcf4882786f",
			Driver: "acpi",
			Data:   map[string]interface{}{},
			Debug:  false,
		},
		machine.DeviceInfo{
			Name:   "f83531c4-8fac-45c9-98bf-21cd6f9919e5",
			Driver: "apic",
			Data:   map[string]interface{}{},
			Debug:  false,
		},
		machine.DeviceInfo{
			Name:   "e10fe9f3-0c22-48da-9f06-692c0e171386",
			Driver: "pit",
			Data:   map[string]interface{}{},
			Debug:  false,
		},
		machine.DeviceInfo{
			Name:   "2e001b29-ece6-4633-a0eb-00858389cc41",
			Driver: "rtc",
			Data:   map[string]interface{}{},
			Debug:  false,
		},
		machine.DeviceInfo{
			Name:   "c9c99c43-1017-48ea-8140-445ae0eed6e8",
			Driver: "pci-bus",
			Data:   map[string]interface{}{},
			Debug:  false,
		},
		machine.DeviceInfo{
			Name:   "5d07808a-6c81-43bd-b0a7-9a05d4340d74",
			Driver: "user-memory",
			Data: map[string]interface{}{
				"fd": "5", "size": "1073741824",
			},
			Debug: false,
		},
		machine.DeviceInfo{
			Name:   "cffe52b7-b0d4-4865-adbd-db2dcb15fe1f",
			Driver: "virtio-pci-console",
			Data:   map[string]interface{}{},
			Debug:  false,
		},
		machine.DeviceInfo{
			Name:   "57c27e13-57b2-4817-9dfc-922460efcdf4",
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
		machine.DeviceInfo{
			Name: "638b8a9c-65be-4fab-b610-37e5cafad230",

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
	}

	vcpus := []platform.VcpuInfo{
		platform.VcpuInfo{
			Id: (*uint)(nil),
			Registers: platform.Registers{
				RAX:       (*platform.RegisterValue)(nil),
				RBX:       (*platform.RegisterValue)(nil),
				RCX:       (*platform.RegisterValue)(nil),
				RDX:       (*platform.RegisterValue)(nil),
				RSI:       (*platform.RegisterValue)(nil),
				RDI:       (*platform.RegisterValue)(nil),
				RSP:       (*platform.RegisterValue)(nil),
				RBP:       (*platform.RegisterValue)(nil),
				R8:        (*platform.RegisterValue)(nil),
				R9:        (*platform.RegisterValue)(nil),
				R10:       (*platform.RegisterValue)(nil),
				R11:       (*platform.RegisterValue)(nil),
				R12:       (*platform.RegisterValue)(nil),
				R13:       (*platform.RegisterValue)(nil),
				R14:       (*platform.RegisterValue)(nil),
				R15:       (*platform.RegisterValue)(nil),
				RIP:       (*platform.RegisterValue)(nil),
				RFLAGS:    (*platform.RegisterValue)(nil),
				CR0:       (*platform.ControlRegisterValue)(nil),
				CR2:       (*platform.ControlRegisterValue)(nil),
				CR3:       (*platform.ControlRegisterValue)(nil),
				CR4:       (*platform.ControlRegisterValue)(nil),
				CR8:       (*platform.ControlRegisterValue)(nil),
				EFER:      (*platform.ControlRegisterValue)(nil),
				APIC_BASE: (*platform.ControlRegisterValue)(nil),
				IDT:       (*platform.DescriptorValue)(nil),
				GDT:       (*platform.DescriptorValue)(nil),
				CS:        (*platform.SegmentValue)(nil),
				DS:        (*platform.SegmentValue)(nil),
				ES:        (*platform.SegmentValue)(nil),
				FS:        (*platform.SegmentValue)(nil),
				GS:        (*platform.SegmentValue)(nil),
				SS:        (*platform.SegmentValue)(nil),
				TR:        (*platform.SegmentValue)(nil),
				LDT:       (*platform.SegmentValue)(nil),
			},
			MpState: (*platform.MpState)(nil),
			Cpuid:   []platform.Cpuid(nil),
			LApic: platform.LApicState{

				Data: []uint8(nil),
			},
			Msrs: []platform.Msr(nil),
			Events: platform.Events{
				Exception:  (*platform.ExceptionEvent)(nil),
				Interrupt:  (*platform.InterruptEvent)(nil),
				NmiPending: false,
				NmiMasked:  false,
				SipiVector: 0x0,
				Flags:      0x0,
			},
			Fpu:   (*platform.Fpu)(nil),
			Xcrs:  []platform.Xcr(nil),
			XSave: (*platform.XSave)(nil),
		},
	}

	state := control.State{
		Devices: devices,
		Vcpus:   vcpus,
	}
	return state
}
