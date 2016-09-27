package cmd

import (
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"syscall"

	"github.com/sevlyar/go-daemon"

	"github.com/rocksolidlabs/novm/src/novm/control"
	"github.com/rocksolidlabs/novm/src/novm/utils"

	"github.com/spf13/cobra"
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "-- Run a new instance.",
	Long: `Run a new instance.

Network definitions are provided as --nic [opt=val],...

    Available options are:

    mac=00:11:22:33:44:55 Set the MAC address.
    tapname=tap1          Set the tap name.
    bridge=br0            Enslave to a bridge.
    ip=192.168.1.2/24     Set the IP address.
    gateway=192.168.1.1   Set the gateway IP.
    debug=true            Enable debugging.

Disk definitions are provided as --disk [opt=val],...

    Available options are:

    filename=disk         Set the backing file (raw).
    dev=vda               Set the device name.
    debug=true            Enable debugging.

Read definitions are provided as a mapping.

    vm_path=>path         Map the given path for reads.

Write definitions are also provided as a mapping.

    vm_path=>path         Map the given path for writes.

    Note that these is always an implicit write path,
    which is a temporary directory for the instance.

        /=>temp_dir

Docker repositories are provided as follows.

    <repository[:tag]>[,key=value]

    You should specify at least username, password.`,
	Run: func(cmd *cobra.Command, args []string) {
		create(cmd)
	},
}

func create(cmd *cobra.Command) {

	context := &daemon.Context{
		PidFileName: "pid",
		PidFilePerm: 0644,
		LogFileName: "log",
		LogFilePerm: 0640,
		WorkDir:     "./",
		Umask:       027,
		Args:        []string{"[novm create ...]"},
	}
	_ = context

	debugCmd(cmd)
	name := flagString(cmd, "name")
	kernel := flagString(cmd, "kernel")
	availableKernels := kernels.List()
	if kernel == "" {
		if availableKernels == nil {
			panic("No kernel found.")
		}
		kernel = availableKernels[0]
	} else {
		if !utils.StringIsIn(kernel, availableKernels) {
			log.Panicf("Kernel '%s' not found!", kernel)
		}
	}
	raw, _ := ioutil.ReadFile(kernels.File(kernel, "release"))
	release := strings.TrimSpace(string(raw))

	cmdline := flagString(cmd, "cmdline")
	if cmdline != "" {
		cmdline += " "
	}
	cpus := flagInt(cmd, "cpus")
	memsize := flagInt(cmd, "memsize")
	init := flagBool(cmd, "init")
	nics := flagStringArray(cmd, "nic")
	disks := flagStringArray(cmd, "disk")
	packs := flagStringSlice(cmd, "pack")
	repos := flagStringArray(cmd, "repo")
	read := flagStringSlice(cmd, "read")
	write := flagStringSlice(cmd, "write")
	vmmopt := flagStringSlice(cmd, "vmmopt")
	cwd := flagString(cmd, "cwd")
	nopci := flagBool(cmd, "nopci")
	com1 := flagBool(cmd, "com1")
	com2 := flagBool(cmd, "com2")
	nofork := flagBool(cmd, "nofork")
	terminal := flagBool(cmd, "terminal")

	_ = cwd
	_ = nofork
	_ = terminal

	cmdArgs := make([]string, 0, 10)
	if init {
		cmdArgs = append(cmdArgs, "-init")
	}

	cmdArgs = append(cmdArgs,
		"-vmlinux", kernels.File(kernel, "vmlinux"),
		"-sysmap", kernels.File(kernel, "sysmap"),
		"-initrd", kernels.File(kernel, "initrd"),
		"-setup", kernels.File(kernel, "setup"))

	// Provide our state by dumping to a temporary file:
	tf, _ := ioutil.TempFile("", "state")
	if debug {
		log.Print("Sate stored in: ", tf.Name())
	} else {
		os.Remove(tf.Name())
	}
	devCmdline, meta := CreateInstance(
		tf, kernel, release, cpus, com1, com2, nopci, memsize,
		nics, disks, read, write, packs, repos, name, cmdline)
	tf.Seek(0, 0)

	// Keep the descriptor:
	stateFd, _ := syscall.Dup(int(tf.Fd()))
	utils.ClearCloexec(stateFd)
	tf.Close()

	// Control socket
	ctrlFileName := path.Join(controlPath, strconv.Itoa(os.Getpid())+".ctrl")
	if debug {
		log.Print("Control file name: ", ctrlFileName)
	}
	ctrlFd, _ := control.Bind(ctrlFileName)

	// Add our metadata to the regitry:
	instances.Add(strconv.Itoa(os.Getpid()), meta)

	cmdArgs = append([]string{
		"-statefd=" + strconv.Itoa(stateFd),
		"-controlfd=" + strconv.Itoa(int(ctrlFd)),
	}, cmdArgs...)

	if devCmdline != "-cmdline=" {
		cmdArgs = append(cmdArgs, devCmdline)
	}

	for _, o := range vmmopt {
		cmdArgs = append(cmdArgs, "-"+o)
	}

	novmm := utils.LibExec("novmm")
	if debug {
		cmdArgs = append(cmdArgs, "-debug")
		log.Print("Executing: ", novmm, " ", strings.Join(cmdArgs, " "))
	}

	command := exec.Command(novmm, cmdArgs...)
	err := command.Run()
	if err != nil {
		log.Panic(err)
	}
}

func init() {
	RootCmd.AddCommand(createCmd)
	createCmd.Flags().String("name", "", "The instance `NAME`.")
	createCmd.Flags().Int("cpus", 1, "The number of `VCPUS`.")
	createCmd.Flags().Int("memsize", 1024, "The memory size `MEMSIZE` (in MB).")
	createCmd.Flags().String("kernel", "", "The `KERNEL` to use.")
	createCmd.Flags().Bool("init", false, "Use a real init?")
	createCmd.Flags().StringArray("nic", nil, "Define a network device (`NIC`).")
	createCmd.Flags().StringArray("disk", nil, "Define a block device (`DISK`).")
	createCmd.Flags().StringSlice("pack", nil, "Use a given read `PACK`.")
	createCmd.Flags().StringArray("repo", nil, "Use a docker repository `REPO`.")
	createCmd.Flags().StringSlice("read", []string{"/"}, "Define a backing filesystem `READ` tree.")
	createCmd.Flags().StringSlice("write", nil, "Define a backing filesystem `WRITE` tree.")
	createCmd.Flags().Bool("nopci", false, "Disable PCI devices?")
	createCmd.Flags().Bool("com1", false, "Enable COM1 UART?")
	createCmd.Flags().Bool("com2", false, "Enable COM2 UART?")
	createCmd.Flags().String("cmdline", "", "Extra command line `CMDLINE` options?")
	createCmd.Flags().StringSlice("vmmopt", nil, "Options to pass to novmm `VMMOPT`.")
	createCmd.Flags().Bool("nofork", false, "Don't fork into the background.")
	createCmd.Flags().StringSlice("env", nil, "Specify an environment variable `ENV`.")
	createCmd.Flags().String("cwd", "", "The process working directory `CWD`.")
	createCmd.Flags().Bool("terminal", false, "Change the terminal mode.")
}
