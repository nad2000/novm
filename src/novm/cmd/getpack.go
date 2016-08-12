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
					panic("Missing URL.")
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
