package cmd_test

import (
	"testing"

	log "github.com/Sirupsen/logrus"

	"github.com/rocksolidlabs/novm/src/novm/cmd"
)

func inti() {
	cmd.SetRootDir("/tmp/TEST/")
}

func TestVMCreation(t *testing.T) {

	log.SetLevel(log.DebugLevel)
	_, instanceID, cmdline := cmd.CreateInstance(
		"KERNEL", "RELEASE", 4, true, true, false, 0, nil, nil, nil, nil, nil, nil, nil, "NAME", "CMDLINE")

	t.Log("ID:", instanceID)
	t.Log("CMDLINE:", cmdline)
}

const SAMPLE_STATE string = `{
    "devices": [
        {
            "cmdline": null,
            "data": {},
            "debug": false,
            "driver": "bios",
            "name": "36bb9417-21aa-4d80-8ffa-29290e10697a"
        },
        {
            "cmdline": "intel_pstate=disable",
            "data": {},
            "debug": false,
            "driver": "acpi",
            "name": "ee3068ec-bfa4-43d0-b33b-4bcf4882786f"
        },
        {
            "cmdline": null,
            "data": {},
            "debug": false,
            "driver": "apic",
            "name": "f83531c4-8fac-45c9-98bf-21cd6f9919e5"
        },
        {
            "cmdline": null,
            "data": {},
            "debug": false,
            "driver": "pit",
            "name": "e10fe9f3-0c22-48da-9f06-692c0e171386"
        },
        {
            "cmdline": null,
            "data": {},
            "debug": false,
            "driver": "rtc",
            "name": "2e001b29-ece6-4633-a0eb-00858389cc41"
        },
        {
            "cmdline": "pci=conf1",
            "data": {},
            "debug": false,
            "driver": "pci-bus",
            "name": "c9c99c43-1017-48ea-8140-445ae0eed6e8"
        },
        {
            "cmdline": null,
            "data": {
                "fd": 5,
                "size": 1073741824
            },
            "debug": false,
            "driver": "user-memory",
            "name": "5d07808a-6c81-43bd-b0a7-9a05d4340d74"
        },
        {
            "cmdline": null,
            "data": {},
            "debug": false,
            "driver": "virtio-pci-console",
            "name": "cffe52b7-b0d4-4865-adbd-db2dcb15fe1f"
        },
        {
            "cmdline": null,
            "data": {
                "fdlimit": 0,
                "read": {
                    "/": [
                        "/"
                    ],
                    "/lib/modules/3.19.0-32-generic": [
                        "/home/nad2000/.novm/kernels/172c0b9d54445b91da234acca4f3beb31240f795/modules"
                    ]
                },
                "tag": "root",
                "write": {
                    "/": "/home/nad2000/.novm/instances/14185"
                }
            },
            "debug": false,
            "driver": "virtio-pci-fs",
            "name": "57c27e13-57b2-4817-9dfc-922460efcdf4"
        },
        {
            "cmdline": null,
            "data": {
                "fdlimit": 0,
                "read": {
                    "/": [],
                    "/init": [
                        "/home/nad2000/novm/bin/noguest"
                    ]
                },
                "tag": "init",
                "write": {
                    "/": "/tmp/tmpTi8NDV"
                }
            },
            "debug": false,
            "driver": "virtio-pci-fs",
            "name": "638b8a9c-65be-4fab-b610-37e5cafad230"
        }
    ],
    "vcpus": [
        {}
    ]
}
`
