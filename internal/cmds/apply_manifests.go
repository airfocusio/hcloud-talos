package cmds

import (
	_ "embed"
	"flag"

	"github.com/airfocusio/hcloud-talos/internal"
	"github.com/airfocusio/hcloud-talos/internal/utils"
)

var (
	//go:embed apply_manifests_hcloud_cloud_controller_manager.yaml.tmpl
	hcloudCloudControllerManagerManifestTmpl string
	//go:embed apply_manifests_hcloud_csi_driver.yaml.tmpl
	hcloudCsiDriverManifestTmpl string
)

type ApplyManifestsCommandId struct{}

func (cmdId *ApplyManifestsCommandId) Name() string {
	return "apply-manifests"
}

func (cmdId *ApplyManifestsCommandId) Create() Command {
	return &ApplyManifestsCommand{}
}

type ApplyManifestsCommand struct {
	SkipHcloudCloudControllerManager bool
	SkipHcloudCsiDriver              bool
}

func (cmd *ApplyManifestsCommand) RegisterOpts(flags *flag.FlagSet) {
	flags.BoolVar(&cmd.SkipHcloudCloudControllerManager, "skip-hcloud-cloud-controller-manager", false, "")
	flags.BoolVar(&cmd.SkipHcloudCsiDriver, "skip-hcloud-csi-driver", false, "")
}

func (cmd *ApplyManifestsCommand) ValidateOpts() error {
	return nil
}

func (cmd *ApplyManifestsCommand) Run(logger *utils.Logger, dir string) error {
	ctx := &internal.Context{Dir: dir}
	err := ctx.Load(logger)
	if err != nil {
		return err
	}

	network, err := internal.HcloudEnsureNetwork(ctx)
	if err != nil {
		return err
	}

	hcloudCloudControllerManagerImage := "hetznercloud/hcloud-cloud-controller-manager:v1.9.1"
	hcloudCloudControllerManagerManifest, err := utils.RenderTemplate(hcloudCloudControllerManagerManifestTmpl, map[string]interface{}{
		"Secret": map[string]interface{}{
			"Token":   ctx.Config.Hcloud.Token,
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
			"Token": ctx.Config.Hcloud.Token,
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
	if !cmd.SkipHcloudCloudControllerManager {
		manifestsConcatenated = append(manifestsConcatenated, []byte(hcloudCloudControllerManagerManifest))
	}
	if !cmd.SkipHcloudCsiDriver {
		manifestsConcatenated = append(manifestsConcatenated, []byte(hcloudCsiDriverManifest))
	}

	manifests, err := utils.YamlSplitMany(manifestsConcatenated...)
	if err != nil {
		return err
	}
	for _, manifest := range manifests {
		err = utils.Retry(ctx.Logger, func() error {
			return internal.KubernetesCreateFromManifest(ctx, string(manifest))
		})
		if err != nil {
			return err
		}
	}

	return nil
}
