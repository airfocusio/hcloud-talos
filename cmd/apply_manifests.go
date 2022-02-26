package cmd

import (
	"github.com/airfocusio/hcloud-talos/internal"
	"github.com/airfocusio/hcloud-talos/internal/utils"
	"github.com/spf13/cobra"
)

var (
	applyManifestsCmdNoHcloudCloudControllerManager bool
	applyManifestsCmdNoHcloudCsiDriver              bool
	applyManifestsCmd                               = &cobra.Command{
		Use:   "apply-manifests",
		Short: "Apply manifests",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			logger := utils.NewLogger(verbose)
			err := internal.ApplyManifests(&logger, dir, internal.ApplyManifestsOpts{
				NoHcloudCloudControllerManager: applyManifestsCmdNoHcloudCloudControllerManager,
				NoHcloudCsiDriver:              applyManifestsCmdNoHcloudCsiDriver,
			})
			return err
		},
	}
)

func init() {
	applyManifestsCmd.Flags().BoolVar(&applyManifestsCmdNoHcloudCloudControllerManager, "no-hcloud-cloud-controller-manager", false, "")
	applyManifestsCmd.Flags().BoolVar(&applyManifestsCmdNoHcloudCsiDriver, "no-hcloud-csi-driver", false, "")
}
