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
	"fmt"

	"github.com/spf13/cobra"
)

// controlCmd represents the control command
var controlCmd = &cobra.Command{
	Use:   "control",
	Short: "-- Execute a control command.",
	Long: `Execute a control command.

Available commands depend on the VMM.

For example:

	To pause the first VCPU:

		vcpu id=0 paused=true

	To enable tracing:

		trace enable=true`,
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: Work your own magic here
		fmt.Println("control called")
	},
}

func init() {
	RootCmd.AddCommand(controlCmd)
	controlCmd.Flags().String("id", "", "a virtual machine `ID`.")
	controlCmd.Flags().String("name", "", "a virtual machine `NAME`.")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// controlCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// controlCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}
