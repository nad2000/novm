package cmd

import (
	"fmt"
	"os"
	"path"

	log "github.com/Sirupsen/logrus"
	"github.com/rocksolidlabs/novm/src/novm/utils"

	"github.com/spf13/cobra"
)

// mkpackageCmd represents the mkpack command
var mkpackageCmd = &cobra.Command{
	Use:   "mkpackage",
	Short: "-- Create a packaage from a tree.",
	Long:  `Create a package from a tree.`,
	Run: func(cmd *cobra.Command, args []string) {
		debugCmd(cmd)
		name := flagString(cmd, "name")
		appdir := flagString(cmd, "appdir")
		include := flagStringSlice(cmd, "include")
		exclude := flagStringSlice(cmd, "exclude")
		mount := flagString(cmd, "mount")
		output := flagString(cmd, "output")

		if output == "" {
			output = utils.TempFileName("tmp", ".pkg")
		}

		if appdir == "" {
			log.Fatal("Missing application directory.")
		}

		if mount == "" {
			log.Fatal("Missing application mount point.")
		}

		dest := utils.TempFileName("", "package")
		defer os.RemoveAll(dest) // clean up

		// filter - "filter function"
		var filter func(string) (bool, error)
		if include != nil && exclude != nil && len(include) > 0 && len(exclude) > 0 {
			filter = utils.CreateFilter(include, exclude)
		}

		err := utils.CopyTree(appdir, path.Join(dest, "ROOT"), filter, true)
		if err != nil {
			log.Fatal("Failed to prepare a copy of the packcage in the destination directory %q: %s", dest, err.Error())
		}

		meta, err := os.Create(path.Join(dest, "PACKAGE.json"))
		fmt.Fprintf(meta, `{
		"name" %q,
		"mount": %q,
		"appdir": %q
}`, name, mount, appdir)
		meta.Close()

		err = utils.Tazdir(dest, output, nil, nil)

		if err != nil {
			log.Fatal(err)
		} else {
			printResult(cmd, "file://"+output)
		}
	},
}

func init() {
	RootCmd.AddCommand(mkpackageCmd)
	mkpackageCmd.Flags().String("name", "", "The package `NAME`.")
	mkpackageCmd.Flags().String("output", "", "The `OUTPUT` file.")
	mkpackageCmd.Flags().String("appdir", "", "The input `APPDIR`.")
	mkpackageCmd.Flags().String("mount", "", "The package mount point on the VM `MOUNT`.")
	mkpackageCmd.Flags().StringSlice("exclude", nil, "Subpaths to `EXCLUDE`.")
	mkpackageCmd.Flags().StringSlice("include", nil, "Subpaths to `INCLUDE`.")
}
