package cmd

import "github.com/spf13/cobra"

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "-- List running instances.",
	Long:  `List running instances.`,
	Run: func(cmd *cobra.Command, args []string) {
		debugCmd(cmd)
		List(flagBool(cmd, "full"), flagBool(cmd, "alive"))
	},
}

// List lists all instances:
func List(full bool, alive bool) {
	rval := instances.Show()
	for pid, v := range rval {
		if isAlive(pid) {
			v["alive"] = "True"
		} else {
			v["alive"] = "False"
		}
	}
	PrettyPrintFields(rval,
		[]string{"id", "name", "timestamp", "cpus", "alive", "ips", "memory"})
}

func init() {
	RootCmd.AddCommand(listCmd)
	listCmd.Flags().Bool("full", false, "Include device info?")
	listCmd.Flags().Bool("alive", false, "Include only alive instances?")
}
