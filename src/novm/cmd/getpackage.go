package cmd

import (
	log "github.com/Sirupsen/logrus"

	"github.com/spf13/cobra"
)

// getpackageCmd represents the getpack command
var getpackageCmd = &cobra.Command{
	Use:   "getpackage",
	Short: "-- Fetch a new package.",
	Long:  `Fetch a new package.`,
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
			result = packages.Fetch(url, dict{"name": name})
			printResult(cmd, result)
		}
	},
}

func init() {
	RootCmd.AddCommand(getpackageCmd)
	getpackageCmd.Flags().Bool("nocache", false, "Don't use a cached version.")
	getpackageCmd.Flags().String("url", "", "Package `URL`.")
	getpackageCmd.Flags().String("name", "", "Package `NAME`")
}
