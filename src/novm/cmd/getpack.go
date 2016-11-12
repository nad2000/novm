package cmd

import (
	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
)

// getpackCmd represents the getpack command
var getpackCmd = &cobra.Command{
	Use:   "getpack",
	Short: "-- Fetch a new pack.",
	Long:  `Fetch a new pack.`,
	Run: func(cmd *cobra.Command, args []string) {
		debugCmd(cmd)
		url := flagString(cmd, "url")
		name := flagString(cmd, "name")
		if result := packs.Find("", dict{"url": url, "name": name}); result != "" {
			printResult(cmd, result)
		} else {
			if url == "" {
				if len(args) > 0 {
					url = args[0]
				} else {
					log.Fatal("Missing URL.")
					return
				}
			}
			result = packs.Fetch(url, dict{"name": name})
			printResult(cmd, result)
		}
	},
}

func init() {
	RootCmd.AddCommand(getpackCmd)
	getpackCmd.Flags().Bool("nocache", false, "Don't use a cached version.")
	getpackCmd.Flags().String("url", "", "Pack `URL`.")
	getpackCmd.Flags().String("name", "", "Pack `NAME`")
}
