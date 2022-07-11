package cmd

import (
	"github.com/airfocusio/hcloud-talos/internal"
	"github.com/airfocusio/hcloud-talos/internal/utils"
	"github.com/spf13/cobra"
)

var (
	destroyClusterCmdConfigFile string
	destroyClusterCmdForce      bool
	destroyClusterCmd           = &cobra.Command{
		Use:   "destroy-cluster",
		Short: "Destroy the cluster",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			logger := utils.NewLogger(verbose)
			err := internal.DestroyCluster(&logger, dir, internal.DestroyClusterOpts{
				ConfigFile: destroyClusterCmdConfigFile,
				Force:      destroyClusterCmdForce,
			})
			return err
		},
	}
)

func init() {
	destroyClusterCmd.Flags().StringVarP(&destroyClusterCmdConfigFile, "config", "c", defaultConfigFile, "")
	destroyClusterCmd.Flags().BoolVar(&destroyClusterCmdForce, "force", false, "")
}
