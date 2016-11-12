package cmd

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"os"
	"path"
	"strings"

	"github.com/rocksolidlabs/novm/src/novm/nodb"
	"github.com/rocksolidlabs/novm/src/novm/utils"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// DB root directory
	root                 string
	controlPath, cfgFile string
	instances, packs     nodb.Nodb
	kernels, dockerDB    nodb.Nodb
	debug                bool
	verbose              bool
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "novm [-h] [--debug] [--plain] [--json] [--verbose]",
	Short: "'novm' management",
	Long: `novm is a legacy-free, type 2 hypervisor written in Go. Its goal is to provide an alternate, high-performance Linux hypervisor for cloud workloads.

novm is unique because it exposes a filesystem-device as a primary mechanism for running guests. This allows you to easily manage independent software and data bundles independently and combine them into a single virtual machine instance.

novm leverages the excellent Linux Kernel Virtual Machine (KVM) interface to run guest instances.`,
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func debugCmd(cmd *cobra.Command) {
	debug = flagBool(cmd, "debug")
	verbose = flagBool(cmd, "verbose")
	utils.Debug = debug
	if debug {
		title := fmt.Sprintf("Command %q called with flags:", cmd.Name())
		fmt.Println(title)
		fmt.Println(strings.Repeat("=", len(title)))
		cmd.DebugFlags()
	}
}

// SetRootDir sets the root directory of the configuration.
// It could be specified using NOVM_ROOT environment variable
func SetRootDir(dirName string) {
	if dirName == "" {
		root = utils.GetEnv(
			"NOVM_ROOT",
			path.Join(utils.GetEnv("HOME", os.TempDir()), ".novm"))
	} else {
		root = dirName
	}
	controlPath = path.Join(root, "control")
	instances.Open(sysDir("NOVM_INSTANCES", "instances"))
	packs.Open(sysDir("NOVM_PACKS", "packs"))
	kernels.Open(sysDir("NOVM_KERNELS", "kernels"))
	dockerDB.Open(sysDir("NOVM_DOCKER", "docker"))
}

func init() {

	SetRootDir("")
	cobra.OnInitialize(initConfig)

	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "`CONFIG` file (default: $HOME/.novm/.config)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	RootCmd.PersistentFlags().BoolP("toggle", "t", false, "Help message for toggle")
	RootCmd.PersistentFlags().BoolP("debug", "d", false, "Show full stack trace on error.")
	RootCmd.PersistentFlags().BoolP("verbose", "V", false, "Show more verbose details.")
	RootCmd.PersistentFlags().BoolP("plain", "p", false, "Print result as plain text (where applicable).")
	RootCmd.PersistentFlags().BoolP("json", "j", false, "Print result as JSON (where applicable).")
}

func sysDir(envVar string, dirName string) string {
	return utils.GetEnv(envVar, path.Join(root, dirName))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" { // enable ability to specify config file via flag
		viper.SetConfigFile(cfgFile)
	}

	viper.SetConfigName(".config") // name of config file (without extension)
	viper.AddConfigPath(root)      // adding home directory as first search path
	viper.AutomaticEnv()           // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

func flagString(cmd *cobra.Command, name string) string {
	return cmd.Flag(name).Value.String()
}

func flagStringSlice(cmd *cobra.Command, name string) (val []string) {
	val, err := cmd.Flags().GetStringSlice(name)
	if err != nil {
		log.Fatal(err)
	}
	return
}

func flagStringArray(cmd *cobra.Command, name string) (val []string) {
	val, err := cmd.Flags().GetStringArray(name)
	if err != nil {
		log.Fatal(err)
	}
	return
}

func flagBool(cmd *cobra.Command, name string) (val bool) {
	val, err := cmd.Flags().GetBool(name)
	if err != nil {
		log.Fatal(err)
	}
	return
}

func flagInt(cmd *cobra.Command, name string) (val int) {
	val, err := cmd.Flags().GetInt(name)
	if err != nil {
		log.Fatal(err)
	}
	return
}

// Print the result
func printResult(cmd *cobra.Command, result interface{}) {
	if json, _ := cmd.Flags().GetBool("json"); json || cmd.Name() == "control" {
		fmt.Println(utils.JSONDumps(result))
	} else if plain, _ := cmd.Flags().GetBool("plain"); plain {
		fmt.Printf("%v\n", result)
	} else {
		fmt.Printf("%+v\n", result)
	}
}

// Is this process still around?
func isAlive(pid interface{}) bool {
	switch pid.(type) {
	default:
		log.Panicf("Incorrect PID type: %#v", pid)
		return false
	case int:
		return utils.PathExists(fmt.Sprintf("/proc/%d", pid.(int)))
	case string:
		return utils.PathExists(fmt.Sprintf("/proc/%s", pid.(string)))
	}
}
