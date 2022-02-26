package cmd

import (
	"github.com/airfocusio/hcloud-talos/internal"
	"github.com/airfocusio/hcloud-talos/internal/utils"
	"github.com/spf13/cobra"
)

var (
	reconcilePoolCmdServerType     string
	reconcilePoolCmdNodeNamePrefix string
	reconcilePoolCmdNodeCount      int
	reconcilePoolCmdForce          bool
	reconcilePoolCmd               = &cobra.Command{
		Use:   "reconcile-pool [pool-name]",
		Short: "Reconcile pool",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			logger := utils.NewLogger(verbose)
			err := internal.ReconcilePool(&logger, dir, internal.ReconcilePoolOpts{
				NodeNamePrefix: reconcilePoolCmdNodeNamePrefix,
				NodeCount:      reconcilePoolCmdNodeCount,
				ServerType:     reconcilePoolCmdServerType,
				Force:          reconcilePoolCmdForce,
				PoolName:       args[0],
			})
			return err
		},
	}
)

func init() {
	reconcilePoolCmd.Flags().BoolVar(&reconcilePoolCmdForce, "force", false, "")
	reconcilePoolCmd.Flags().StringVar(&reconcilePoolCmdServerType, "server-type", "cx21", "")
	reconcilePoolCmd.Flags().StringVar(&reconcilePoolCmdNodeNamePrefix, "node-name-prefix", "worker", "")
	reconcilePoolCmd.Flags().IntVar(&reconcilePoolCmdNodeCount, "node-count", 1, "")
}
