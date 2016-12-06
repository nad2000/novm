package cmd

import (
	"os"
	"path"
	"strconv"
	"syscall"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/spf13/cobra"
)

// stopCmd represents the stop command
var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "-- Stop a running VM.",
	Long:  `Stop a running VM.`,
	Run: func(cmd *cobra.Command, args []string) {

		debugCmd(cmd)

		objID, pidStr := instanceIDAndPid(cmd)

		if objID == "" {
			if len(args) < 1 {
				log.Fatal("Missing instance ID, PID, or name")
			}
			objID = instances.Find(args[0], nil)
			if objID == "" {
				pid, err := strconv.Atoi(args[0])
				if err != nil {
					objID = instances.Find("", dict{"name": args[0]})
					if objID == "" {
						log.Fatal("Failed to find instance.")
					}
				}
				objID = instances.FindIDbyPid(pid)
			}
		}

		if pidStr == "" {
			pidStr = InstancePid(objID)
		}

		if !isAlive(pidStr) {
			log.Fatalf("Instace %q is not running (PID: %s).", objID, pidStr)
		}

		log.Info("Sending SIGTERM to VM process ", pidStr)
		pid, _ := strconv.Atoi(pidStr)
		syscall.Kill(pid, syscall.SIGTERM)
		// waiting untill it gets terminated (max 20 sec)
		waits := 0
		for waits < 20 {
			time.Sleep(time.Second)
			if !isAlive(pidStr) {
				goto Done
			}
			waits += 1
			os.Stderr.WriteString(".")
		}

		if isAlive(pidStr) {
			syscall.Kill(pid, syscall.SIGKILL)
		}

	Done:

		if waits > 0 {
			os.Stderr.WriteString("\n")
		}

		if flagBool(cmd, "rm") {
			if instances.Remove(objID, nil) {
				if pidStr != "" {
					os.RemoveAll(path.Join(controlPath, objID+".ctrl"))
				}
				log.Printf("The instance %q removed.", objID)
			}
		}
	},
}

func init() {
	RootCmd.AddCommand(stopCmd)
	stopCmd.Flags().String("id", "", "The instance `ID`.")
	stopCmd.Flags().Int("pid", 0, "The instance `PID`.")
	stopCmd.Flags().String("name", "", "The instance `NAME`.")
	stopCmd.Flags().Bool("rm", false, "Remove the instance after the termination.")
}
