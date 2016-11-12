package cmd

import (
	log "github.com/Sirupsen/logrus"
	"os"
	"path"

	"github.com/spf13/cobra"
)

// cleanCmd represents the clean command
var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "-- Remove stale instance information.",
	Long:  `Remove stale instance information.`,
	Run: func(cmd *cobra.Command, args []string) {
		debugCmd(cmd)
		id, name := flagString(cmd, "id"), flagString(cmd, "name")
		id = instances.Find(id, dict{"name": name})
		if instances.Remove(id, nil) {
			os.RemoveAll(path.Join(controlPath, id+".ctrl"))
			log.Printf("The instance %q removed.", name+id)
		}
	},
}

func init() {
	RootCmd.AddCommand(cleanCmd)
	cleanCmd.Flags().String("id", "", "The instance `ID`.")
	cleanCmd.Flags().String("name", "", "The instance `NAME`.")
}
