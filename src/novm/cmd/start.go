package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"strconv"
	"strings"
	"syscall"

	log "github.com/Sirupsen/logrus"

	"github.com/rocksolidlabs/novm/src/novm/control"
	"github.com/rocksolidlabs/novm/src/novm/utils"

	"github.com/spf13/cobra"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "-- Start a stopped instance.",
	Long:  `Start a stopped instance.`,
	Run: func(cmd *cobra.Command, args []string) {

		nofork := flagBool(cmd, "nofork")

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
		start(cmd)
	},
}

func start(cmd *cobra.Command) {

	signals := make(chan os.Signal, 2)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-signals
		utils.Cleanup()
		os.Exit(0)
	}()

	debugCmd(cmd)
	name := flagString(cmd, "name")
	id := flagString(cmd, "id")
	remove := flagBool(cmd, "rm")

	instanceID := instances.Find(id, map[string]string{"name": name})
	if instanceID == "" {
		log.Fatalf("Specified instance ID=%q, name=%q doesn't exit", id, name)
	}

	config := GetInstanceConfig(instanceID)

	cmdArgs := make([]string, 0, 10)
	if config.Init {
		cmdArgs = append(cmdArgs, "-init")
	}

	kernel := config.Kernel
	cmdArgs = append(cmdArgs,
		"-vmlinux", kernels.File(kernel, "vmlinux"),
		"-sysmap", kernels.File(kernel, "sysmap"),
		"-initrd", kernels.File(kernel, "initrd"),
		"-setup", kernels.File(kernel, "setup"))

	stateFd, devCmdline := ActivateInstance(instanceID, config, remove)

	// Control socket
	ctrlFileName := path.Join(controlPath, strconv.Itoa(os.Getpid())+".ctrl")
	if debug {
		log.Print("Control file name: ", ctrlFileName)
	}
	ctrlFd, _ := control.Bind(ctrlFileName)

	utils.AddForCleaning(instances.File(instanceID) + ".json")
	utils.AddForCleaning(instances.File(instanceID))

	cmdArgs = append(cmdArgs,
		"-statefd="+strconv.Itoa(stateFd),
		"-controlfd="+strconv.Itoa(int(ctrlFd)))

	// ignore empty "-cmdline"
	if devCmdline != "-cmdline=" {
		cmdArgs = append(cmdArgs, devCmdline)
	}

	// for _, o := range vmmopt {
	// 	cmdArgs = append(cmdArgs, "-"+o)
	// }

	novmm := utils.LibExec("novmm")
	if debug {
		if verbose {
			cmdArgs = append(cmdArgs, "-debug")
		}
		log.Print("Executing: ", novmm, " ", strings.Join(cmdArgs[0:], " "))
	}

	if debug {
		fmt.Print("Press 'Enter' to continue...")
		bufio.NewReader(os.Stdin).ReadBytes('\n')
	}

	command := exec.Command(novmm, cmdArgs...)
	command.Stdin = os.Stdin
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	command.SysProcAttr = &syscall.SysProcAttr{}
	command.SysProcAttr.Credential = &syscall.Credential{Uid: uint32(0), Gid: uint32(0)}
	err := command.Run()

	utils.Cleanup()
	if err != nil {
		log.Print(err)
		os.Exit(1)
	}
}

func init() {
	RootCmd.AddCommand(startCmd)
	startCmd.Flags().String("name", "", "The instance `NAME`.")
	startCmd.Flags().String("id", "", "The instance `ID`.")
	startCmd.Flags().Bool("nofork", false, "Don't fork into the background.")
	startCmd.Flags().Bool("rm", false, "Remove the instance after the termination.")
}
