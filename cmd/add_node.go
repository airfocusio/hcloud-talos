package cmd

import (
	"github.com/airfocusio/hcloud-talos/internal"
	"github.com/airfocusio/hcloud-talos/internal/utils"
	"github.com/spf13/cobra"
)

var (
	addNodeCmdConfigFile   string
	addNodeCmdControlplane bool
	addNodeCmdServerType   string
	addNodeCmdPoolName     string
	addNodeCmdTalosVersion string
	addNodeCmd             = &cobra.Command{
		Use:   "add-node [node-name]",
		Short: "Add a new node",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			logger := utils.NewLogger(verbose)
			_, err := internal.AddNode(&logger, dir, internal.AddNodeOpts{
				ConfigFile:   addNodeCmdConfigFile,
				ServerType:   addNodeCmdServerType,
				Controlplane: addNodeCmdControlplane,
				PoolName:     addNodeCmdPoolName,
				NodeName:     args[0],
				TalosVersion: addNodeCmdTalosVersion,
			})
			return err
		},
	}
)

func init() {
	addNodeCmd.Flags().StringVarP(&addNodeCmdConfigFile, "config", "c", defaultConfigFile, "")
	addNodeCmd.Flags().BoolVar(&addNodeCmdControlplane, "controlplane", false, "")
	addNodeCmd.Flags().StringVar(&addNodeCmdServerType, "server-type", "cx21", "")
	addNodeCmd.Flags().StringVar(&addNodeCmdPoolName, "pool-name", "", "")
	addNodeCmd.Flags().StringVar(&addNodeCmdTalosVersion, "talos-version", "", "")
}
