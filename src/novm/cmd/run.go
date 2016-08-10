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

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "-- Execute a command inside a novm.",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.DebugFlags()
		fmt.Printf("run called with %#v", args)
	},
}

func init() {
	RootCmd.AddCommand(runCmd)
	runCmd.Flags().Bool("terminal", false, "Change the terminal mode.")
	runCmd.Flags().String("id", "", "a virtual machine `ID`.")
	runCmd.Flags().String("name", "", "a virtual machine `NAME`.")
	runCmd.Flags().String("env", "", "specify an environment variable `ENV`.")
	runCmd.Flags().String("cwd", "", "current working directory `CWD`.")
}
