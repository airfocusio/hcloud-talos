package cmd

import (
	"github.com/airfocusio/hcloud-talos/internal"
	"github.com/airfocusio/hcloud-talos/internal/utils"
	"github.com/spf13/cobra"
)

var (
	applyManifestsCmdConfigFile                     string
	applyManifestsCmdNoFlannel                      bool
	applyManifestsCmdNoHcloudCloudControllerManager bool
	applyManifestsCmdNoHcloudCsiDriver              bool
	applyManifestsCmd                               = &cobra.Command{
		Use:   "apply-manifests",
		Short: "Apply manifests",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			logger := utils.NewLogger(verbose)
			err := internal.ApplyManifests(&logger, dir, internal.ApplyManifestsOpts{
				ConfigFile:                     applyManifestsCmdConfigFile,
				NoFlannel:                      applyManifestsCmdNoFlannel,
				NoHcloudCloudControllerManager: applyManifestsCmdNoHcloudCloudControllerManager,
				NoHcloudCsiDriver:              applyManifestsCmdNoHcloudCsiDriver,
			})
			return err
		},
	}
)

func init() {
	applyManifestsCmd.Flags().StringVarP(&applyManifestsCmdConfigFile, "config", "c", defaultConfigFile, "")
	applyManifestsCmd.Flags().BoolVar(&applyManifestsCmdNoFlannel, "no-flannel", false, "")
	applyManifestsCmd.Flags().BoolVar(&applyManifestsCmdNoHcloudCloudControllerManager, "no-hcloud-cloud-controller-manager", false, "")
	applyManifestsCmd.Flags().BoolVar(&applyManifestsCmdNoHcloudCsiDriver, "no-hcloud-csi-driver", false, "")
}
