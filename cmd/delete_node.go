package cmd

import (
	"github.com/airfocusio/hcloud-talos/internal"
	"github.com/airfocusio/hcloud-talos/internal/utils"
	"github.com/spf13/cobra"
)

var (
	deleteNodeCmdKeepServer bool
	deleteNodeCmdForce      bool
	deleteNodeCmd           = &cobra.Command{
		Use:   "delete-node [node-name]",
		Short: "Delete a node",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			logger := utils.NewLogger(verbose)
			err := internal.DeleteNode(&logger, dir, internal.DeleteNodeOpts{
				KeepServer: deleteNodeCmdKeepServer,
				Force:      deleteNodeCmdForce,
				NodeName:   args[0],
			})
			return err
		},
	}
)

func init() {
	deleteNodeCmd.Flags().BoolVar(&deleteNodeCmdKeepServer, "keep-server", false, "")
	deleteNodeCmd.Flags().BoolVar(&deleteNodeCmdForce, "force", false, "")
}
