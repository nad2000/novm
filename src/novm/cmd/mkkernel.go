package cmd

import (
	"io"
	"io/ioutil"
	"log"
	"novm/utils"
	"os"
	"os/exec"
	"path"
	"path/filepath"

	"github.com/spf13/cobra"
)

// mkkernelCmd represents the mkkernel command
var mkkernelCmd = &cobra.Command{
	Use:   "mkkernel",
	Short: "-- Make a new kernel from an local kernel.",
	Long:  `Make a new kernel from an local kernel.`,
	Run: func(cmd *cobra.Command, args []string) {

		debugCmd(cmd)
		output := flagString(cmd, "output")
		release := flagString(cmd, "release")
		modules := flagString(cmd, "modules")
		bzimage := flagString(cmd, "bzimage")
		vmlinux := flagString(cmd, "vmlinux")
		setup := flagString(cmd, "setup")
		sysmap := flagString(cmd, "sysmap")
		nomodules := flagBool(cmd, "nomodules")
		result := Mkkernel(output, release, modules, bzimage, vmlinux, setup, sysmap, nomodules)
		printResult(cmd, result)
	},
}

func Mkkernel(

	output string, release string, modules string, bzimage string,
	vmlinux string, setup string, sysmap string, nomodules bool) string {
	if output == "" {
		output = utils.TempFileName("", ".ker")
	}
	if release == "" {
		release = utils.SystemRelease()
	}
	if modules == "" {
		modules = "/lib/modules/" + release
	}
	if bzimage == "" {
		bzimage = "/boot/vmlinuz-" + release
	}
	if vmlinux == "" {
		vmlinuxFile, err := ioutil.TempFile("", "ker")
		if err != nil {
			log.Fatal(err)
		}

		command := exec.Command(utils.LibExec("extract-vmlinux"), bzimage)
		command.Stdout = vmlinuxFile

		if debug {
			log.Printf("Runing 'extract-vmlinux' with bzimage=%q to %q\n",
				bzimage, vmlinuxFile.Name())
		}

		if err := command.Run(); err != nil {
			log.Panicln(err)
		}
		if err := vmlinuxFile.Close(); err != nil {
			log.Panicln(err)
		}
		vmlinux = vmlinuxFile.Name()

	}
	if setup == "" {
		setupFile, _ := ioutil.TempFile("", "")
		bzimageFile, _ := os.Open(bzimage)
		io.Copy(setupFile, bzimageFile)
		bzimageFile.Close()
		setup = setupFile.Name()
		setupFile.Close()
	}

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

	if nomodules {
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

	absOutput, _ := filepath.Abs(path.Clean(output))
	return "file://" + absOutput
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
