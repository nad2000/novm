package cmd

import (
	"novm/utils"

	"github.com/spf13/cobra"
)

// kernelsCmd represents the kernels command
var kernelsCmd = &cobra.Command{
	Use:   "kernels",
	Short: "-- List available kernels.",
	Long:  `List available kernels.`,
	Run: func(cmd *cobra.Command, args []string) {
		debugCmd(cmd)
		Kernels()
	},
}

func Kernels() {
	utils.PrettyPrintFields(kernels.Show(),
		[]string{"id", "url", "timestamp", "release", "name"})
}

func init() {
	RootCmd.AddCommand(kernelsCmd)
}
