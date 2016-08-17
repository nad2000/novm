package cmd

import "github.com/spf13/cobra"

// rmkernelCmd represents the rmkernel command
var rmkernelCmd = &cobra.Command{
	Use:   "rmkernel",
	Short: "-- Remove an existing kernel.",
	Long:  `Remove an existing kernel.`,
	Run: func(cmd *cobra.Command, args []string) {
		debugCmd(cmd)
		id := flagString(cmd, "id")
		url := flagString(cmd, "url")
		name := flagString(cmd, "name")
		kernels.Remove(id, dict{"url": url, "name": name})
	},
}

func init() {
	RootCmd.AddCommand(rmkernelCmd)
	rmkernelCmd.Flags().String("id", "", "The kernel `ID`.")
	rmkernelCmd.Flags().String("name", "", "The kernel `NAME`.")
	rmkernelCmd.Flags().String("url", "", "The kernel `URL`")
}
