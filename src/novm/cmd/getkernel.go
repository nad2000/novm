package cmd

import (
	log "github.com/Sirupsen/logrus"
	"os"

	"github.com/spf13/cobra"
)

// getkernelCmd represents the getkernel command
var getkernelCmd = &cobra.Command{
	Use:   "getkernel",
	Short: "-- Fetch a new kernel.",
	Long:  `Fetch a new kernel.`,
	Run: func(cmd *cobra.Command, args []string) {
		debugCmd(cmd)
		url := flagString(cmd, "url")
		nocache := flagBool(cmd, "nocache")
		name := flagString(cmd, "name")
		if url == "" {
			if len(args) > 0 {
				url = args[0]
			} else {
				log.Fatal("Missing kernel URL.")
				os.Exit(1)
			}
		}
		printResult(cmd, GetKernel(url, nocache, name))
	},
}

// GetKernel attempts to import a kernel with given URL naming as 'name'
func GetKernel(url string, nocache bool, name string) interface{} {
	_ = nocache
	if result := kernels.Find("", dict{"url": url, "name": name}); result != "" {
		return result
	}

	return kernels.Fetch(url, dict{"name": name})
}

func init() {
	RootCmd.AddCommand(getkernelCmd)
	getkernelCmd.Flags().String("url", "", "The kernel `URL` (e.g. file: or http:).")
	getkernelCmd.Flags().Bool("nocache", false, "Don't use a cached version.")
	getkernelCmd.Flags().String("name", "", "A user-provided `NAME`.")
}
