package cmd

import (
	log "github.com/Sirupsen/logrus"

	"github.com/rocksolidlabs/novm/src/novm/docker"

	"github.com/spf13/cobra"
)

// getimageCmd represents the getimage command
var getimageCmd = &cobra.Command{
	Use:   "getimage",
	Short: "-- Download a docker image.",
	Long: `Download the specified docker image in the form:
	
	<repository[:tag]>[,key=value]

    With key and value pairs you can specify username, password, docker registry host etc.`,
	Run: func(cmd *cobra.Command, args []string) {
		debugCmd(cmd)
		if len(args) < 1 {
			log.Fatalf("Missing docker images specification.")
		}
		docker.PullImages(dockerDB, args)
	},
}

func init() {
	RootCmd.AddCommand(getimageCmd)
}
