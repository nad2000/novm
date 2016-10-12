package cmd

import (
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"

	"github.com/spf13/cobra"
)

// cleanallCmd represents the cleanall command
var cleanallCmd = &cobra.Command{
	Use:   "cleanall",
	Short: "-- Remove everything not alive.",
	Long:  `Remove everything not alive.`,
	Run: func(cmd *cobra.Command, args []string) {
		debugCmd(cmd)
		for _, pid := range instances.List() {
			if !isAlive(pid) {
				if debug {
					log.Print("Removing PID: ", pid)
				}
				instances.Remove(pid, nil)
				os.RemoveAll(path.Join(controlPath, pid+".ctrl"))
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
