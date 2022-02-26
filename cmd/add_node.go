package cmd

import (
	"github.com/airfocusio/hcloud-talos/internal"
	"github.com/airfocusio/hcloud-talos/internal/utils"
	"github.com/spf13/cobra"
)

var (
	addNodeCmdControlplane bool
	addNodeCmdServerType   string
	addNodeCmdPoolName     string
	addNodeCmd             = &cobra.Command{
		Use:   "add-node [node-name]",
		Short: "Add a new node",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			logger := utils.NewLogger(verbose)
			_, err := internal.AddNode(&logger, dir, internal.AddNodeOpts{
				ServerType:   addNodeCmdServerType,
				Controlplane: addNodeCmdControlplane,
				PoolName:     addNodeCmdPoolName,
				NodeName:     args[0],
			})
			return err
		},
	}
)

func init() {
	addNodeCmd.Flags().BoolVar(&addNodeCmdControlplane, "controlplane", false, "")
	addNodeCmd.Flags().StringVar(&addNodeCmdServerType, "server-type", "cx21", "")
	addNodeCmd.Flags().StringVar(&addNodeCmdPoolName, "pool-name", "", "")
}
