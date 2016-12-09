package cmd

import (
	"io/ioutil"
	"os"
	"path"
	"strings"

	log "github.com/Sirupsen/logrus"

	"github.com/spf13/cobra"
)

// cleanallCmd represents the cleanall command
var cleanallCmd = &cobra.Command{
	Use:   "cleanall",
	Short: "-- Remove everything not alive.",
	Long:  `Remove everything not alive.`,
	Run: func(cmd *cobra.Command, args []string) {
		debugCmd(cmd)
		for _, id := range instances.List() {

			pid := InstancePid(id)
			if !isAlive(pid) || pid == "" {
				if debug {
					if pid != "" {
						log.Print("Removing PID: ", pid)
					} else {
						log.Print("Removing instace: ", id)
					}
				}
				instances.Remove(id, nil)
				if pid != "" {
					os.RemoveAll(path.Join(controlPath, pid+".ctrl"))
				}
			}
		}
		files, _ := ioutil.ReadDir(controlPath)
		for _, fi := range files {
			pid := strings.TrimSuffix(fi.Name(), ".ctrl")
			if !isAlive(pid) {
				os.RemoveAll(path.Join(controlPath, fi.Name()))
			}
		}
	},
}

func init() {
	RootCmd.AddCommand(cleanallCmd)
}
