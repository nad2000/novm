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
				panic("Missing kernel URL.")
				return
			}
		}
		printResult(cmd, GetKernel(url, nocache, name))
	},
}

func GetKernel(url string, nocache bool, name string) interface{} {
	_ = nocache
	if result := kernels.Find("", dict{"url": url, "name": name}); result != "" {
		return result
	} else {
		return kernels.Fetch(url, dict{"name": name})
	}
}

func init() {
	RootCmd.AddCommand(getkernelCmd)
	getkernelCmd.Flags().String("url", "", "The kernel `URL` (e.g. file: or http:).")
	getkernelCmd.Flags().Bool("nocache", false, "Don't use a cached version.")
	getkernelCmd.Flags().String("name", "", "A user-provided `NAME`.")
}
