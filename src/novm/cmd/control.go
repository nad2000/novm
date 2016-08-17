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
	"log"
	"novm/control"
	"path"
	"strings"

	"github.com/spf13/cobra"
)

type dict map[string]string

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
		cmd.DebugFlags()
		id := flagString(cmd, "id")
		name := flagString(cmd, "name")
		if args == nil {
			log.Fatal("Invalid command. Expected a VMM command.")
			return
		}
		command := args[0]
		params := make(map[string]string)
		for _, arg := range args[1:] {
			splitArg := strings.Split(arg, "=")
			if len(splitArg) != 2 {
				log.Fatalf("Arguments should be key=value instead of %v.", arg)
				return
			}
			params[splitArg[0]] = splitArg[1]
		}

		objId := instances.Find(id, dict{"name": name})
		ctrlPath := path.Join(control_file, objId+".ctrl")
		control.Setup(ctrlPath)
		fmt.Printf("control called: %#v, %#v, %#v, %#v", name, objId, params, command)
		result, err := control.Rpc(command, params)
		fmt.Printf("Result: %#v, %#v", result, err)

	},
}

func init() {
	RootCmd.AddCommand(controlCmd)
	controlCmd.Flags().String("id", "", "a virtual machine `ID`.")
	controlCmd.Flags().String("name", "", "The instance `NAME`.")
}
