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
	"os"
	"path"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	root    string // DB root directory
	cfgFile string
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
	//	Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {

	root = path.Join(getEnv("HOME", os.TempDir()), ".novm")
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports Persistent Flags, which, if defined here,
	// will be global for your application.

	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "`CONFIG` file (default is $HOME/.novm.yaml)")
	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	RootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	RootCmd.Flags().BoolP("debug", "d", false, "Show full stack trace on error.")
	RootCmd.Flags().BoolP("plain", "p", false, "Print result as plain text (where applicable).")
	RootCmd.Flags().BoolP("json", "j", false, "Print result as JSON (where applicable).")
}

func getEnv(name string, defaultValue string) string {
	// ruturns environent variable value or it's missing, default fall back value

	value := os.Getenv(name)
	if value == "" {
		return defaultValue
	}
	return value
}

func getDbDir(envVar string, dirName string) string {
	return getEnv(envVar, path.Join(root, dirName))
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
