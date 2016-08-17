package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "-- Execute a command inside a novm.",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.DebugFlags()
		fmt.Printf("run called with %#v", args)
	},
}

func init() {
	RootCmd.AddCommand(runCmd)
	runCmd.Flags().Bool("terminal", false, "Change the terminal mode.")
	runCmd.Flags().String("id", "", "a virtual machine `ID`.")
	runCmd.Flags().String("name", "", "a virtual machine `NAME`.")
	runCmd.Flags().String("env", "", "specify an environment variable `ENV`.")
	runCmd.Flags().String("cwd", "", "current working directory `CWD`.")
}
