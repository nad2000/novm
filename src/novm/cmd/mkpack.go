package cmd

import (
	log "github.com/Sirupsen/logrus"
	"github.com/rocksolidlabs/novm/src/novm/utils"

	"github.com/spf13/cobra"
)

// mkpackCmd represents the mkpack command
var mkpackCmd = &cobra.Command{
	Use:   "mkpack",
	Short: "-- Create a pack from a tree.",
	Long:  `Create a pack from a tree.`,
	Run: func(cmd *cobra.Command, args []string) {
		debugCmd(cmd)
		objID := flagString(cmd, "id")
		name := flagString(cmd, "name")
		dirPath := flagString(cmd, "path")
		include := flagStringSlice(cmd, "include")
		exclude := flagStringSlice(cmd, "exclude")
		output := utils.TempFileName("tmp", ".tar.gz")

		if dirPath == "" {
			if objID != "" || name != "" {
				objID = instances.Find(objID, dict{"name": name})
				dirPath = instances.File(objID)
			} else {
				dirPath = "."
			}
		}

		err := utils.Tazdir(dirPath, output, include, exclude)
		if err != nil {
			log.Fatal(err)
		} else {
			printResult(cmd, "file://"+output)
		}
	},
}

func init() {
	RootCmd.AddCommand(mkpackCmd)
	mkpackCmd.Flags().String("id", "", "The instance `ID`.")
	mkpackCmd.Flags().String("name", "", "The instance `NAME`.")
	mkpackCmd.Flags().String("output", "", "The `OUTPUT` file.")
	mkpackCmd.Flags().String("path", "", "The input `PATH`.")
	mkpackCmd.Flags().StringSlice("exclude", nil, "Subpaths to `EXCLUDE`.")
	mkpackCmd.Flags().StringSlice("include", nil, "Subpaths to `INCLUDE`.")
}
