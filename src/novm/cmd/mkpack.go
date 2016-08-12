// Copyright © 2016 Radomirs Cirskis <nad2000@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"novm/utils"

	"github.com/spf13/cobra"
)

// mkpackCmd represents the mkpack command
var mkpackCmd = &cobra.Command{
	Use:   "mkpack",
	Short: "-- Create a pack from a tree.",
	Long:  `Create a pack from a tree.`,
	Run: func(cmd *cobra.Command, args []string) {
		debugCmd(cmd)
		objId := flagString(cmd, "id")
		name := flagString(cmd, "name")
		dirPath := flagString(cmd, "path")
		include := flagStringSlice(cmd, "include")
		exclude := flagStringSlice(cmd, "exclude")
		output := utils.TempFileName("tmp", ".zip")

		if dirPath == "" {
			if objId != "" || name != "" {
				objId = instances.Find(objId, dict{"name": name})
				dirPath = instances.File(objId)
			} else {
				dirPath = "."
			}
		}

		err := utils.Packdir(dirPath, output, include, exclude)
		if err != nil {
			panic(err)
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
