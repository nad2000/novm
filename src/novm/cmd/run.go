package cmd

import (
	"log"
	"github.com/rocksolidlabs/novm/src/novm/control"
	"path"

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
		id := flagString(cmd, "id")
		name := flagString(cmd, "name")
		env := flagStringSlice(cmd, "env")
		cwd := flagString(cmd, "cwd")
		command := args[:]

		if command == nil || len(command) == 0 {
			log.Fatal("Invalid command.")
			return
		}

		objId := instances.Find(id, dict{"name": name})
		ctrlPath := path.Join(control_file, objId+".ctrl")
		control.Open(ctrlPath)

		result, _ := control.Run(command, env, cwd, terminal)
		printResult(cmd, result)
	},
}

func init() {
	RootCmd.AddCommand(runCmd)
	runCmd.Flags().Bool("terminal", false, "Change the terminal mode.")
	runCmd.Flags().String("id", "", "The instance`ID`.")
	runCmd.Flags().String("name", "", "The instance `NAME`.")
	runCmd.Flags().StringSlice("env", nil, "Specify an environment variable `ENV`.")
	runCmd.Flags().String("cwd", "", "The process working directory (`CWD`).")
}
