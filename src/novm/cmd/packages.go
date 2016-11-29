package cmd

import (
	"github.com/rocksolidlabs/novm/src/novm/utils"

	"github.com/spf13/cobra"
)

var packagesCmd = &cobra.Command{
	Use:   "packages",
	Short: "-- List available packages.",
	Long:  `List available packages.`,
	Run: func(cmd *cobra.Command, args []string) {
		debugCmd(cmd)
		Packages()
	},
}

// Packages lists available packs
func Packages() {
	rval := packages.Show()
	utils.PrettyPrintFields(rval,
		[]string{"id", "url", "timestamp", "name", "appdir", "mount"})
}

func init() {
	RootCmd.AddCommand(packagesCmd)
}
