package cmd

import (
	"github.com/rocksolidlabs/novm/src/novm/utils"

	"github.com/spf13/cobra"
)

var packsCmd = &cobra.Command{
	Use:   "packs",
	Short: "-- List available packs.",
	Long:  `List available packs.`,
	Run: func(cmd *cobra.Command, args []string) {
		debugCmd(cmd)
		Packs()
	},
}

// Packs lists available packs
func Packs() {
	rval := packs.Show()
	utils.PrettyPrintFields(rval,
		[]string{"id", "url", "timestamp", "name"})
}

func init() {
	RootCmd.AddCommand(packsCmd)
}
