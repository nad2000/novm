package cmd

import (
	"os"
	"path"
	"strconv"

	log "github.com/Sirupsen/logrus"

	"github.com/rocksolidlabs/novm/src/novm/control"

	"github.com/spf13/cobra"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "-- Execute a command inside a novm.",
	Long:  `Execute a command inside a novm.`,
	Run: func(cmd *cobra.Command, args []string) {

		var objID string

		debugCmd(cmd)
		terminal := flagBool(cmd, "terminal")
		id := flagString(cmd, "id")
		pid := flagInt(cmd, "pid")
		name := flagString(cmd, "name")
		env := flagStringSlice(cmd, "env")
		cwd := flagString(cmd, "cwd")
		command := args[:]

		if command == nil || len(command) == 0 {
			log.Fatal("Invalid command.")
			return
		}

		if pid > 0 {
			objID = instances.Find(id, dict{"pid": strconv.Itoa(pid)})
		} else {
			objID = instances.Find(id, dict{"name": name})
		}

		if objID == "" {
			if pid > 0 {
				log.Fatalf("Instance with PID %d doesn't exit.", pid)
			} else {
				log.Fatalf("Instance with ID %q doesn't exit.", id)
			}
		}

		instance, _ := instances.Get(objID)
		pidStr, ok := instance["pid"]
		if !ok {
			log.Fatalf("Missing PID of the instance %q.", objID)
		}
		if !isAlive(pidStr) {
			log.Fatalf("Instace %d is not running (PID: %s).", objID, pidStr)
		}

		ctrlPath := path.Join(controlPath, pidStr+".ctrl")
		control.Open(ctrlPath)
		defer control.Close()

		result, err, exitCode := control.Run(command, env, cwd, terminal)
		if err != nil {
			log.Print(err)
		}
		printResult(cmd, result)
		if exitCode != 0 {
			os.Exit(exitCode)
		}
	},
}

func init() {
	RootCmd.AddCommand(runCmd)
	runCmd.Flags().Bool("terminal", false, "Change the terminal mode.")
	runCmd.Flags().String("id", "", "The instance `ID`.")
	runCmd.Flags().Int("pid", 0, "The instance `PID`.")
	runCmd.Flags().String("name", "", "The instance `NAME`.")
	runCmd.Flags().StringSlice("env", nil, "Specify an environment variable `ENV`.")
	runCmd.Flags().String("cwd", "", "The process working directory (`CWD`).")
}
