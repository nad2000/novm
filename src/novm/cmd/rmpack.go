package cmd

import "github.com/spf13/cobra"

// rmpackCmd represents the rmpack command
var rmpackCmd = &cobra.Command{
	Use:   "rmpack",
	Short: "-- Remove an existing pack.",
	Long:  `Remove an existing pack.`,
	Run: func(cmd *cobra.Command, args []string) {
		debugCmd(cmd)
		id := flagString(cmd, "id")
		url := flagString(cmd, "url")
		name := flagString(cmd, "name")
		packs.Remove(id, dict{"url": url, "name": name})
	},
}

func init() {
	RootCmd.AddCommand(rmpackCmd)
	rmpackCmd.Flags().String("id", "", "The pack `ID`.")
	rmpackCmd.Flags().String("name", "", "The pack `NAME`.")
	rmpackCmd.Flags().String("url", "", "The pack `URL`")
}
