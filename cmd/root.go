package cmd

import (
	"github.com/spf13/cobra"
)

var (
	verbose bool
	dir     string
	rootCmd = &cobra.Command{
		Use: "hcloud-talos",
	}
)

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose")
	rootCmd.PersistentFlags().StringVarP(&dir, "dir", "d", ".", "")
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(addNodeCmd)
	rootCmd.AddCommand(applyManifestsCmd)
	rootCmd.AddCommand(bootstrapClusterCmd)
	rootCmd.AddCommand(deleteNodeCmd)
	rootCmd.AddCommand(destroyClusterCmd)
	rootCmd.AddCommand(reconcilePoolCmd)
}

func Execute() error {
	return rootCmd.Execute()
}
