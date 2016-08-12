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

import "github.com/spf13/cobra"

// rmpackCmd represents the rmpack command
var rmpackCmd = &cobra.Command{
	Use:   "rmpack",
	Short: "-- Remove an existing pack.",
	Long:  `Remove an existing pack.`,
	Run: func(cmd *cobra.Command, args []string) {
		debugCmd(cmd)
		id := flagString(cmd, "id")
		url := flagString(cmd, "url")
		name := flagString(cmd, "name")
		packs.Remove(id, dict{"url": url, "name": name})
	},
}

func init() {
	RootCmd.AddCommand(rmpackCmd)
	rmpackCmd.Flags().String("id", "", "The pack `ID`.")
	rmpackCmd.Flags().String("name", "", "The pack `NAME`.")
	rmpackCmd.Flags().String("url", "", "The pack `URL`")
}
