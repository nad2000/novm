package cmd

import (
	"strconv"
	"syscall"

	log "github.com/Sirupsen/logrus"

	"github.com/rocksolidlabs/novm/src/novm/utils"

	"github.com/spf13/cobra"
)

// cloneCmd represents the clone command
var cloneCmd = &cobra.Command{
	Use:   "clone",
	Short: "-- clone an instance.",
	Long:  `clone an instance.`,
	Run: func(cmd *cobra.Command, args []string) {
		clone(cmd)
	},
}

func clone(cmd *cobra.Command) {

	var pid int

	debugCmd(cmd)
	name := flagString(cmd, "name")
	id := flagString(cmd, "id")
	newName := flagString(cmd, "new-name")
	includeRoot := flagBool(cmd, "root")

	instanceID := instances.Find(id, map[string]string{"name": name})
	if instanceID == "" {
		log.Fatalf("Specified instance ID=%q, name=%q doesn't exit", id, name)
	}

	newID := utils.NewID()
	instance, _ := instances.Get(instanceID)

	if name == "" {
		name = instance["name"]
	}
	// create instace name from the cloned instance name
	if newName == "" && name != "" {
		for i := 1; ; i++ {
			newName = name + "_" + strconv.Itoa(i)
			found := instances.Find(nil, dict{"name": newName})
			if found == "" {
				break
			}
		}
	}

	instance["name"] = newName
	instances.Add(newID, instance)

	config := GetInstanceConfig(instanceID)
	config.Name = newName
	config.Save(newID)

	if includeRoot {
		// Suspend the instance if it's active (to copy the root FS)
		pidStr := InstancePid(instanceID)
		if isAlive(pidStr) {
			pid, _ = strconv.Atoi(pidStr)
			syscall.Kill(pid, syscall.SIGSTOP)
		}

		utils.CopyTree(
			instances.File(instanceID),
			instances.File(newID),
			nil, false)

		// Resume the instance if it was active
		if includeRoot && pid > 0 {
			syscall.Kill(pid, syscall.SIGCONT)
		}
	}
}

func init() {
	RootCmd.AddCommand(cloneCmd)
	cloneCmd.Flags().String("id", "", "The instance `ID`.")
	cloneCmd.Flags().String("name", "", "The instance `NAME`.")
	cloneCmd.Flags().String("new-name", "", "The cloned instance `NEWNAME`.")
	cloneCmd.Flags().BoolP("root", "r", false, "Include the copy of the root system (directory) of the source.")
}
