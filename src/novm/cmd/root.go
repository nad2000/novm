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
	"os"
	"path"
	"strings"

	"novm/nodb"
	"novm/utils"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// DB root directory
	root, controls, cfgFile string
	instances, packs        nodb.Nodb
	kernels, docker         nodb.Nodb
	debug                   bool
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "novm [-h] [--debug] [--plain] [--json]",
	Short: "TODO: A brief description of your application",
	Long: `TODO: A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		debugCmd(cmd)
	},
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
	if debug {
		title := fmt.Sprintf("Command %q called with flags:", cmd.Name())
		fmt.Println(title)
		fmt.Println(strings.Repeat("=", len(title)))
		cmd.DebugFlags()
	}
}

func init() {

	root = utils.GetEnv(
		"NOVM_ROOT",
		path.Join(utils.GetEnv("HOME", os.TempDir()), ".novm"))
	controls = path.Join(root, "controls")
	instances.Open(sysDir("NOVM_INSTANCES", "instances"))
	packs.Open(sysDir("NOVM_PACKS", "packs"))
	kernels.Open(sysDir("NOVM_KERNELS", "kernels"))
	docker.Open(sysDir("NOVM_DOCKER", "docker"))

	cobra.OnInitialize(initConfig)

	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "`CONFIG` file (default: $HOME/.novm/.config)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	RootCmd.PersistentFlags().BoolP("toggle", "t", false, "Help message for toggle")
	RootCmd.PersistentFlags().BoolP("debug", "d", false, "Show full stack trace on error.")
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
		fmt.Println(utils.JsonDumps(result))
	} else if plain, _ := cmd.Flags().GetBool("plain"); plain {
		fmt.Printf("%v", result)
	} else {
		fmt.Printf("%+v", result)
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
