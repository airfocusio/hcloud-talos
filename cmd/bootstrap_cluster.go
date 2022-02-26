package cmd

import (
	"os"

	"github.com/airfocusio/hcloud-talos/internal"
	"github.com/airfocusio/hcloud-talos/internal/utils"
	"github.com/spf13/cobra"
)

var (
	bootstrapClusterCmdServerType                     string
	bootstrapClusterCmdLocation                       string
	bootstrapClusterCmdNetworkZone                    string
	bootstrapClusterCmdNoFirewall                     bool
	bootstrapClusterCmdNoTalosKubespan                bool
	bootstrapClusterCmdNoHcloudCloudControllerManager bool
	bootstrapClusterCmdNoHcloudCsiDriver              bool
	bootstrapClusterCmd                               = &cobra.Command{
		Use:   "bootstrap-cluster [cluster-name] [node-name]",
		Short: "Bootstrap a new cluster",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			logger := utils.NewLogger(verbose)
			err := internal.BootstrapCluster(&logger, dir, internal.BootstrapClusterOpts{
				ClusterName:                    args[0],
				NodeName:                       args[1],
				ServerType:                     bootstrapClusterCmdServerType,
				Location:                       bootstrapClusterCmdLocation,
				NetworkZone:                    bootstrapClusterCmdNetworkZone,
				Token:                          os.Getenv("HCLOUD_TOKEN"),
				NoFirewall:                     bootstrapClusterCmdNoFirewall,
				NoTalosKubespan:                bootstrapClusterCmdNoTalosKubespan,
				NoHcloudCloudControllerManager: bootstrapClusterCmdNoHcloudCloudControllerManager,
				NoHcloudCsiDriver:              bootstrapClusterCmdNoHcloudCsiDriver,
			})
			return err
		},
	}
)

func init() {
	bootstrapClusterCmd.Flags().StringVar(&bootstrapClusterCmdServerType, "server-type", "cx21", "")
	bootstrapClusterCmd.Flags().StringVar(&bootstrapClusterCmdLocation, "location", "nbg1", "")
	bootstrapClusterCmd.Flags().StringVar(&bootstrapClusterCmdNetworkZone, "network-zone", "eu-central", "")
	bootstrapClusterCmd.Flags().BoolVar(&bootstrapClusterCmdNoFirewall, "no-firewall", false, "")
	bootstrapClusterCmd.Flags().BoolVar(&bootstrapClusterCmdNoTalosKubespan, "no-talos-kubespan", false, "")
	bootstrapClusterCmd.Flags().BoolVar(&bootstrapClusterCmdNoHcloudCloudControllerManager, "no-hcloud-cloud-controller-manager", false, "")
	bootstrapClusterCmd.Flags().BoolVar(&bootstrapClusterCmdNoHcloudCsiDriver, "no-hcloud-csi-driver", false, "")
}
