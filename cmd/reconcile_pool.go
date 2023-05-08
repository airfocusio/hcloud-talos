package cmd

import (
	"github.com/airfocusio/hcloud-talos/internal"
	"github.com/airfocusio/hcloud-talos/internal/utils"
	"github.com/spf13/cobra"
)

var (
	reconcilePoolCmdConfigFile     string
	reconcilePoolCmdServerType     string
	reconcilePoolCmdNodeNamePrefix string
	reconcilePoolCmdNodeCount      int
	reconcilePoolCmdTalosVersion   string
	reconcilePoolCmd               = &cobra.Command{
		Use:   "reconcile-pool [pool-name]",
		Short: "Reconcile pool",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			logger := utils.NewLogger(verbose)
			err := internal.ReconcilePool(&logger, dir, internal.ReconcilePoolOpts{
				ConfigFile:     reconcilePoolCmdConfigFile,
				NodeNamePrefix: reconcilePoolCmdNodeNamePrefix,
				NodeCount:      reconcilePoolCmdNodeCount,
				ServerType:     reconcilePoolCmdServerType,
				PoolName:       args[0],
				TalosVersion:   reconcilePoolCmdTalosVersion,
			})
			return err
		},
	}
)

func init() {
	reconcilePoolCmd.Flags().StringVarP(&reconcilePoolCmdConfigFile, "config", "c", defaultConfigFile, "")
	reconcilePoolCmd.Flags().StringVar(&reconcilePoolCmdServerType, "server-type", "cx21", "")
	reconcilePoolCmd.Flags().StringVar(&reconcilePoolCmdNodeNamePrefix, "node-name-prefix", "worker", "")
	reconcilePoolCmd.Flags().IntVar(&reconcilePoolCmdNodeCount, "node-count", 1, "")
	reconcilePoolCmd.Flags().StringVar(&reconcilePoolCmdTalosVersion, "talos-version", "", "")
}
