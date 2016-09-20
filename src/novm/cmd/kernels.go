package cmd

import (
	"bufio"
	"os"
	"path"

	"github.com/rocksolidlabs/novm/src/novm/utils"

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

// Kernels prints the list of all available kernels
func Kernels() {
	items := kernels.Show()
	for objID, item := range items {
		rf, err := os.Open(path.Join(kernels.File(objID), "release"))
		if err != nil {
			continue
		}
		r := bufio.NewReader(rf)
		relaese, _ := r.ReadString(0x0A)
		item["release"] = relaese
		rf.Close()
	}
	utils.PrettyPrintFields(items,
		[]string{"id", "url", "timestamp", "release", "name"})
}

func init() {
	RootCmd.AddCommand(kernelsCmd)
}
