package cmd

import (
	log "github.com/Sirupsen/logrus"
	"path"
	"strings"

	"github.com/rocksolidlabs/novm/src/novm/control"

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

		trace enable=true
`,
	Run: func(cmd *cobra.Command, args []string) {

		debugCmd(cmd)
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

		objID := instances.Find(id, dict{"name": name})
		ctrlPath := path.Join(controlPath, objID+".ctrl")
		control.Open(ctrlPath)
		result, _ := control.RPC(command, params)
		printResult(cmd, result)
	},
}

func init() {
	RootCmd.AddCommand(controlCmd)
	controlCmd.Flags().String("id", "", "a virtual machine `ID`.")
	controlCmd.Flags().String("name", "", "The instance `NAME`.")
}
