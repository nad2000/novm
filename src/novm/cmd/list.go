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
	. "novm/utils"

	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "-- List running instances.",
	Long:  `List running instances.`,
	Run: func(cmd *cobra.Command, args []string) {
		debugCmd(cmd)
		List(flagBool(cmd, "full"), flagBool(cmd, "alive"))
	},
}

func List(full bool, alive bool) {
	rval := instances.Show()
	PrettyPrint(rval)
}

func init() {
	RootCmd.AddCommand(listCmd)
	listCmd.Flags().Bool("full", false, "Include device info?")
	listCmd.Flags().Bool("alive", false, "Include only alive instances?")
}
