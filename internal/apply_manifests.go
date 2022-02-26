package internal

import (
	_ "embed"

	"github.com/airfocusio/hcloud-talos/internal/clients"
	"github.com/airfocusio/hcloud-talos/internal/cluster"
	"github.com/airfocusio/hcloud-talos/internal/utils"
)

var (
	//go:embed apply_manifests_hcloud_cloud_controller_manager.yaml
	hcloudCloudControllerManagerManifestTmpl string
	//go:embed apply_manifests_hcloud_csi_driver.yaml
	hcloudCsiDriverManifestTmpl string
)

type ApplyManifestsOpts struct {
	NoHcloudCloudControllerManager bool
	NoHcloudCsiDriver              bool
}

func ApplyManifests(logger *utils.Logger, dir string, opts ApplyManifestsOpts) error {
	cl := &cluster.Cluster{Dir: dir}
	err := cl.Load(logger)
	if err != nil {
		return err
	}

	network, err := clients.HcloudEnsureNetwork(cl, nodeNetworkTemplate(cl), false)
	if err != nil {
		return err
	}

	hcloudCloudControllerManagerImage := "hetznercloud/hcloud-cloud-controller-manager:v1.9.1"
	hcloudCloudControllerManagerManifest, err := utils.RenderTemplate(hcloudCloudControllerManagerManifestTmpl, map[string]interface{}{
		"Secret": map[string]interface{}{
			"Token":   cl.Config.Hcloud.Token,
			"Network": network.Name,
		},
		"CloudControllerManager": map[string]interface{}{
			"Image":           hcloudCloudControllerManagerImage,
			"ImagePullPolicy": "IfNotPresent",
		},
	})
	if err != nil {
		return err
	}

	hcloudCsiDriverImage := "hetznercloud/hcloud-csi-driver:1.6.0"
	hcloudCsiDriverManifest, err := utils.RenderTemplate(hcloudCsiDriverManifestTmpl, map[string]interface{}{
		"Secret": map[string]interface{}{
			"Token": cl.Config.Hcloud.Token,
		},
		"CsiDriver": map[string]interface{}{
			"Image":           hcloudCsiDriverImage,
			"ImagePullPolicy": "IfNotPresent",
		},
	})
	if err != nil {
		return err
	}

	manifestsConcatenated := [][]byte{}
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
