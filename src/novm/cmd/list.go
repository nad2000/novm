package cmd

import (
	"io/ioutil"

	"github.com/rocksolidlabs/novm/src/novm/utils"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "ls",
	Aliases: []string{"list"},
	Short:   "-- List running instances.",
	Long:    `List running instances.`,
	Run: func(cmd *cobra.Command, args []string) {
		debugCmd(cmd)
		List(flagBool(cmd, "full"), flagBool(cmd, "alive"))
	},
}

// InstancePid returns instace's PID if it is running
func InstancePid(instanceID string) (pidStr string) {
	pidFileName := instances.File("." + instanceID + ".pid")
	if utils.PathExists(pidFileName) {
		pid, _ := ioutil.ReadFile(pidFileName)
		pidStr = string(pid)
	}
	return
}

// List lists all instances:
func List(full bool, alive bool) {
	rval := instances.Show()
	for id, v := range rval {

		if pidStr := InstancePid(id); pidStr != "" && isAlive(pidStr) {
			v["pid"] = pidStr
			v["alive"] = "True"
		} else {
			v["alive"] = "False"
		}
	}
	utils.PrettyPrintFields(rval,
		[]string{"id", "pid", "name", "timestamp", "cpus", "alive", "ips", "memory"})
}

func init() {
	RootCmd.AddCommand(listCmd)
	listCmd.Flags().Bool("full", false, "Include device info?")
	listCmd.Flags().Bool("alive", false, "Include only alive instances?")
}
