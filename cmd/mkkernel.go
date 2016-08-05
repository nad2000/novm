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
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"

	"../utils"

	"github.com/spf13/cobra"
)

// mkkernelCmd represents the mkkernel command
var mkkernelCmd = &cobra.Command{
	Use:   "mkkernel",
	Short: "-- Make a new kernel from an local kernel.",
	Long:  `Make a new kernel from an local kernel.`,
	Run: func(cmd *cobra.Command, args []string) {

		cmd.DebugFlags()

		output := flagString(cmd, "output")
		if output == "" {
			output = utils.TempFileName("", "")
		}
		release := flagString(cmd, "release")
		if release == "" {
			release = utils.SystemRelease()
		}
		modules := flagString(cmd, "modules")
		if modules == "" {
			modules = "/lib/modules/" + release
		}
		bzimage := flagString(cmd, "bzimage")
		if bzimage == "" {
			bzimage = "/boot/vmlinuz-" + release
		}
		vmlinux := flagString(cmd, "vmlinux")
		if vmlinux == "" {
			vmlinuxFile, err := ioutil.TempFile("", "")
			if err != nil {
				log.Fatal(err)
			}

			command := exec.Command(utils.LibExec("extract-vmlinux"), bzimage)
			command.Stdout = vmlinuxFile

			if err := command.Run(); err != nil {
				log.Fatal(err)
			}
			if err := vmlinuxFile.Close(); err != nil {
				log.Fatal(err)
			}

			vmlinux = vmlinuxFile.Name()
		}
		setup := flagString(cmd, "setup")
		if setup == "" {
			setupFile, _ := ioutil.TempFile("", "")
			bzimageFile, _ := os.Open(bzimage)
			io.Copy(setupFile, bzimageFile)
			bzimageFile.Close()
			setup = setupFile.Name()
			setupFile.Close()
		}

		sysmap := flagString(cmd, "sysmap")
		if sysmap == "" {
			sysmap = "/boot/System.map-" + release
		}

		// Copy all the files into a single directory.
		tempDir, _ := ioutil.TempDir(os.TempDir(), "tmp")
		command := exec.Command(utils.LibExec("mkinitramfs"), release, modules)
		initrd, _ := os.Create(path.Join(tempDir, "initrd"))
		command.Stdout = initrd

		if err := command.Run(); err != nil {
			log.Fatal(err)
		}
		if err := initrd.Close(); err != nil {
			log.Fatal(err)
		}
		utils.CopyFile(vmlinux, path.Join(tempDir, "vmlinux"))
		utils.CopyFile(sysmap, path.Join(tempDir, "sysmap"))
		utils.CopyFile(setup, path.Join(tempDir, "setup"))

		if flagBool(cmd, "nomodules") {
			os.MkdirAll(path.Join(tempDir, "modules"), os.ModeDir|0755)
		} else {
			utils.CopyTree(
				modules,
				path.Join(tempDir, "modules"),
				func(name string) bool {
					return name == "source" || name == "build"
				})
		}
		ioutil.WriteFile(path.Join(tempDir, "release"), []byte(release), 0755)
		utils.Packdir(tempDir, output, nil, nil)

		absOutput, _ := filepath.Abs(output)
		printResult(cmd, "file://"+absOutput)

	},
}

func init() {
	RootCmd.AddCommand(mkkernelCmd)

	mkkernelCmd.Flags().String("output", "", "The `OUTPUT` file.")
	mkkernelCmd.Flags().String("release", "", "The kernel `RELEASE` (automated).")
	mkkernelCmd.Flags().String("modules", "", "Path to the kernel `MODULES`.")
	mkkernelCmd.Flags().String("vmlinux", "", "Path to the `VMLINUX` file.")
	mkkernelCmd.Flags().String("bzimage", "", "Path to the compressed image `BZIMAGE`.")
	mkkernelCmd.Flags().String("setup", "", "Path to the `SETUP` header.")
	mkkernelCmd.Flags().String("sysmap", "", "Path to the system map `SYSMAP`.")
	mkkernelCmd.Flags().Bool("nomodules", false, "Don't include modules.")

}
