package cmd

import (
	"encoding/json"
	"log"
	"os"

	"github.com/rocksolidlabs/novm/src/novm/utils"
)

// InstanceConfig stores parameters the instance was created with
type InstanceConfig struct {
	Kernel, Release, Name, Cmdline   string
	Cpus, Memsize                    int
	Com1, Com2, Nopci, Init          bool
	Nics, Disks, Read, Write, Images []string
	Packs, Vmmopt                    []string
	Packages                         []string
}

// Load loads the instance configuraion from JSON configuration file
func (config *InstanceConfig) Load(instanceID string) {
	configFileName := instances.File("." + instanceID + ".config")
	if !utils.PathExists(configFileName) {
		log.Fatal("The instance %q configeter file %q is missing.", instanceID, configFileName)
	}
	cf, _ := os.Open(configFileName)
	json.NewDecoder(cf).Decode(config)
	cf.Close()
}

// GetInstanceConfig retrievs the instace configuration
func GetInstanceConfig(instanceID string) (config InstanceConfig) {
	config.Load(instanceID)
	return
}

// Save serializes and saves instance configuration in JSON configuration file
// a hidden file with extension '.config'
func (config *InstanceConfig) Save(instanceID string) string {
	configFileName := instances.File("." + instanceID + ".config")
	cf, _ := os.Create(configFileName)
	json.NewEncoder(cf).Encode(config)
	cf.Close()
	return configFileName
}
