package internal

import (
	_ "embed"

	"github.com/airfocusio/hcloud-talos/internal/clients"
	"github.com/airfocusio/hcloud-talos/internal/cluster"
	"github.com/airfocusio/hcloud-talos/internal/utils"
)

var (
	//go:embed apply_manifests_hcloud_secret.yaml
	hcloudSecretManifestTmpl string
	//go:embed apply_manifests_hcloud_cloud_controller_manager.yaml
	hcloudCloudControllerManagerManifestTmpl string
	//go:embed apply_manifests_hcloud_csi_driver.yaml
	hcloudCsiDriverManifestTmpl string
)

type ApplyManifestsOpts struct {
	ConfigFile                     string
	NoHcloudCloudControllerManager bool
	NoHcloudCsiDriver              bool
}

func ApplyManifests(logger *utils.Logger, dir string, opts ApplyManifestsOpts) error {
	cl := &cluster.Cluster{Dir: dir}
	err := cl.Load(opts.ConfigFile, logger)
	if err != nil {
		return err
	}

	network, err := clients.HcloudEnsureNetwork(cl, nodeNetworkTemplate(cl), false)
	if err != nil {
		return err
	}

	hcloudSecretManifest, err := utils.RenderTemplate(hcloudSecretManifestTmpl, map[string]interface{}{
		"Token":   cl.Config.Hcloud.Token,
		"Network": network.Name,
	})

	hcloudCloudControllerManagerManifest, err := utils.RenderTemplate(hcloudCloudControllerManagerManifestTmpl, map[string]interface{}{})
	if err != nil {
		return err
	}

	hcloudCsiDriverManifest, err := utils.RenderTemplate(hcloudCsiDriverManifestTmpl, map[string]interface{}{})
	if err != nil {
		return err
	}

	manifestsConcatenated := [][]byte{}
	manifestsConcatenated = append(manifestsConcatenated, []byte(hcloudSecretManifest))
	if !opts.NoHcloudCloudControllerManager {
		manifestsConcatenated = append(manifestsConcatenated, []byte(hcloudCloudControllerManagerManifest))
	}
	if !opts.NoHcloudCsiDriver {
		manifestsConcatenated = append(manifestsConcatenated, []byte(hcloudCsiDriverManifest))
	}

	manifests, err := utils.YamlSplitMany(manifestsConcatenated...)
	if err != nil {
		return err
	}
	for _, manifest := range manifests {
		err = utils.Retry(cl.Logger, func() error {
			return clients.KubernetesCreateFromManifest(cl, string(manifest))
		})
		if err != nil {
			return err
		}
	}

	return nil
}
