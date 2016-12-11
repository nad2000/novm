package cmd

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"syscall"

	log "github.com/Sirupsen/logrus"

	"github.com/cloudfoundry/gosigar/psnotify"

	"github.com/rocksolidlabs/novm/src/novm/control"
	"github.com/rocksolidlabs/novm/src/novm/docker"
	"github.com/rocksolidlabs/novm/src/novm/utils"

	"github.com/spf13/cobra"
)

var dockerImageDirectories []string

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

Docker images are provided as follows.

    <repository[:tag]>[,key=value]

    With key and value pairs you can specify username, password, docker registry host etc.`,
	Run: func(cmd *cobra.Command, args []string) {

		debugCmd(cmd)
		nofork := flagBool(cmd, "nofork")
		images := flagStringArray(cmd, "image")
		if len(images) > 0 {
			dockerImageDirectories = docker.PullImages(dockerDB, images)
		}

		instanceID := utils.NewID()
		if !nofork {
			// detaches the current program by forking and exiting the parent:
			r1, _, err1 := syscall.RawSyscall(syscall.SYS_FORK, 0, 0, 0)
			if err1 != 0 {
				if debug {
					log.Fatal(os.NewSyscallError("fork", error(err1)))
				} else {
					log.Fatalf("Failed to fork %q", os.Args[0])
				}
			}
			//  r1 - child pid = 0 in child
			if r1 != 0 {
				log.Printf("%q PID: %d", os.Args[0], r1)
				os.Exit(0)
				log.Fatal("Exit failed")
			}
			// Reset to NODB root for the child process:
			SetRootDir("")
		}
		create(cmd, instanceID)
	},
}

func create(cmd *cobra.Command, instanceID string) {

	name := flagString(cmd, "name")
	kernel := flagString(cmd, "kernel")
	availableKernels := kernels.List()
	if kernel == "" {
		if availableKernels == nil {
			log.Fatal("No kernel found.")
		}
		kernel = availableKernels[0]
	} else {
		kernel = kernels.Find(kernel, nil)
		if kernel == "" {
			log.Fatalf("Kernel '%s' not found!", kernel)
		}
	}
	raw, _ := ioutil.ReadFile(kernels.File(kernel, "release"))
	release := strings.TrimSpace(string(raw))

	cmdline := flagString(cmd, "cmdline")
	if cmdline != "" {
		cmdline += " "
	}
	terminal := flagBool(cmd, "terminal")
	init := flagBool(cmd, "init")
	vmmopt := flagStringSlice(cmd, "vmmopt")
	cwd := flagString(cmd, "cwd")
	remove := flagBool(cmd, "rm")

	_ = cwd
	_ = terminal

	config := InstanceConfig{
		Name:     name,
		Kernel:   kernel,
		Release:  release,
		Vmmopt:   vmmopt,
		Init:     init,
		Cpus:     flagInt(cmd, "cpus"),
		Com1:     flagBool(cmd, "com1"),
		Com2:     flagBool(cmd, "com2"),
		Nopci:    flagBool(cmd, "nopci"),
		Memsize:  flagInt(cmd, "memsize"),
		Nics:     flagStringArray(cmd, "nic"),
		Disks:    flagStringArray(cmd, "disk"),
		Read:     flagStringSlice(cmd, "read"),
		Write:    flagStringSlice(cmd, "write"),
		Packs:    flagStringSlice(cmd, "pack"),
		Packages: flagStringSlice(cmd, "package"),
		Images:   flagStringArray(cmd, "image"),
		Cmdline:  cmdline}
	cmdArgs := make([]string, 0, 10)
	if init {
		cmdArgs = append(cmdArgs, "-init")
	}

	cmdArgs = append(cmdArgs,
		"-vmlinux", kernels.File(kernel, "vmlinux"),
		"-sysmap", kernels.File(kernel, "sysmap"),
		"-initrd", kernels.File(kernel, "initrd"),
		"-setup", kernels.File(kernel, "setup"))

	stateFd, devCmdline := ActivateInstance(instanceID, config, remove)

	// Control socket
	ctrlFileName := path.Join(controlPath, strconv.Itoa(os.Getpid())+".ctrl")
	//utils.AddForCleaning(ctrlFileName)
	if debug {
		log.Print("Control file name: ", ctrlFileName)
	}
	ctrlFd, _ := control.Bind(ctrlFileName)

	cmdArgs = append(cmdArgs,
		"-statefd="+strconv.Itoa(stateFd),
		"-controlfd="+strconv.Itoa(int(ctrlFd)))

	// ignore empty "-cmdline"
	if devCmdline != "-cmdline=" {
		cmdArgs = append(cmdArgs, devCmdline)
	}

	for _, o := range vmmopt {
		cmdArgs = append(cmdArgs, "-"+o)
	}

	novmm := utils.LibExec("novmm")
	if debug {
		if verbose {
			cmdArgs = append(cmdArgs, "-debug")
		}
		log.Print("Executing: ", novmm, " ", strings.Join(cmdArgs[0:], " "))
	}

	nofork := flagBool(cmd, "nofork")
	if nofork && debug {
		fmt.Print("Press 'Enter' to continue...")
		bufio.NewReader(os.Stdin).ReadBytes('\n')
	}

	command := exec.Command(novmm, cmdArgs...)
	command.Stdin = os.Stdin
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	command.SysProcAttr = &syscall.SysProcAttr{}
	command.SysProcAttr.Credential = &syscall.Credential{Uid: uint32(0), Gid: uint32(0)}

	err := command.Start()
	if err != nil {
		log.Errorf("Error starting NOVMM process: %v", err)
		os.Exit(1)
	}

	watcher, err := psnotify.NewWatcher()
	if err != nil {
		log.Print(err)
		os.Exit(1)
	}

	// Process events
	go func() {
		for {
			select {
			case ev := <-watcher.Exit:
				// cleanup tmp files on exit
				utils.Cleanup()
				fmt.Println("exit event:", ev)
				os.Exit(0)
			case err := <-watcher.Error:
				// cleanup and log errors
				utils.Cleanup()
				fmt.Println("error:", err)
				os.Exit(1)
			}
		}
	}()

	err = watcher.Watch(command.Process.Pid, psnotify.PROC_EVENT_ALL)
	if err != nil {
		log.Errorf("Error watching NOVMM process: %s", err)
		os.Exit(1)
	}

	err = command.Wait()
	if err != nil {
		log.Errorf("Error returned on NOVMM process exit: %s", err)
		os.Exit(1)
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
	createCmd.Flags().StringSlice("package", nil, "Use a given read `PACKAGE`.")
	createCmd.Flags().StringArray("image", nil, "Use a docker image `IMAGE`.")
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
	createCmd.Flags().Bool("rm", false, "Remove the instance after the termination.")
}
