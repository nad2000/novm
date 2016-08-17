package cmd

import "github.com/spf13/cobra"

// cleanallCmd represents the cleanall command
var cleanallCmd = &cobra.Command{
	Use:   "cleanall",
	Short: "-- Remove everything not alive.",
	Long:  `Remove everything not alive.`,
	Run: func(cmd *cobra.Command, args []string) {
		debugCmd(cmd)
		for pid := range instances.List() {
			if isAlive(pid) {
				instances.Remove(pid, nil)
			}
		}
	},
}

func init() {
	RootCmd.AddCommand(cleanallCmd)
}
