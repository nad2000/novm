package cmd

import (
	"os"
	"path"

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

		debugCmd(cmd)
		terminal := flagBool(cmd, "terminal")
		env := flagStringSlice(cmd, "env")
		cwd := flagString(cmd, "cwd")
		command := args[:]

		if command == nil || len(command) == 0 {
			log.Fatal("Invalid command.")
			return
		}

		objID, pidStr := instanceIDAndPid(cmd)
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
