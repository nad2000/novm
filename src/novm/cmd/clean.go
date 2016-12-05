package cmd

import (
	"os"
	"path"

	log "github.com/Sirupsen/logrus"

	"github.com/spf13/cobra"
)

// cleanCmd represents the clean command
var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "-- Remove stale instance information.",
	Long:  `Remove stale instance information.`,
	Run: func(cmd *cobra.Command, args []string) {
		debugCmd(cmd)
		objID, pidStr := instanceIDAndPid(cmd)
		if isAlive(pidStr) {
			log.Fatalf("Instance %q with PID %s is still running.", objID, pidStr)
		}
		if instances.Remove(objID, nil) {
			if pidStr != "" {
				os.RemoveAll(path.Join(controlPath, objID+".ctrl"))
			}
			log.Printf("The instance %q removed.", objID)
		}
	},
}

func init() {
	RootCmd.AddCommand(cleanCmd)
	cleanCmd.Flags().String("id", "", "The instance `ID`.")
	cleanCmd.Flags().Int("pid", 0, "The instance `PID`.")
	cleanCmd.Flags().String("name", "", "The instance `NAME`.")
}
