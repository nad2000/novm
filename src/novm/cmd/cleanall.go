package cmd

import (
	"log"

	"github.com/spf13/cobra"
)

// cleanallCmd represents the cleanall command
var cleanallCmd = &cobra.Command{
	Use:   "cleanall",
	Short: "-- Remove everything not alive.",
	Long:  `Remove everything not alive.`,
	Run: func(cmd *cobra.Command, args []string) {
		debugCmd(cmd)
		for _, pid := range instances.List() {
			log.Print("Removing PID: ", pid)
			if !isAlive(pid) {
				if debug {
					log.Print("Removing PID: ", pid)
				}
				instances.Remove(pid, nil)
			}
		}
	},
}

func init() {
	RootCmd.AddCommand(cleanallCmd)
}
