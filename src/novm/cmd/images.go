package cmd

import (
	"github.com/rocksolidlabs/novm/src/novm/utils"

	"github.com/spf13/cobra"
)

// imagesCmd represents the images command
var imagesCmd = &cobra.Command{
	Use:   "images",
	Short: "-- list all downloaded dockder images.",
	Long:  "-- list all downloaded dockder images.",
	Run: func(cmd *cobra.Command, args []string) {
		debugCmd(cmd)
		rval := dockerDB.Show()
		utils.PrettyPrintFields(rval, []string{"id", "name", "tag", "timestamp"})
	},
}

func init() {
	RootCmd.AddCommand(imagesCmd)
}
