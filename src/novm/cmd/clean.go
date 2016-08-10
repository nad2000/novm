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
	"log"

	"github.com/spf13/cobra"
)

// cleanCmd represents the clean command
var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "-- Remove stale instance information.",
	Long:  `Remove stale instance information.`,
	Run: func(cmd *cobra.Command, args []string) {
		debugCmd(cmd)
		id, name := flagString(cmd, "id"), flagString(cmd, "name")
		if instances.Remove(id, dict{"name": name}) {
			log.Printf("The instance %q removed.", name+id)
		}
	},
}

func init() {
	RootCmd.AddCommand(cleanCmd)
	cleanCmd.Flags().String("id", "", "The instance `ID`.")
	cleanCmd.Flags().String("name", "", "The instance `NAME`.")
}
