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

// kernelsCmd represents the kernels command
var kernelsCmd = &cobra.Command{
	Use:   "kernels",
	Short: "-- List available kernels.",
	Long:  `List available kernels.`,
	Run: func(cmd *cobra.Command, args []string) {
		debugCmd(cmd)
		Kernels()
	},
}

func Kernels() {
	utils.PrettyPrintFields(kernels.Show(),
		[]string{"id", "url", "timestamp", "release", "name"})
}

func init() {
	RootCmd.AddCommand(kernelsCmd)
}
