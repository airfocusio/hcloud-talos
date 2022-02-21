package internal

import (
	_ "embed"
	"fmt"
	"io/ioutil"
	"os/exec"
	"path"
	"strings"

	"github.com/hetznercloud/hcloud-go/hcloud"
)

var (
	//go:embed talos_config_patch.json
	talosConfigPatch string
)

func TalosGenConfig(ctx *Context, clusterName string, controlplaneIP string) error {
	if err := talosCmd(ctx.Dir, "gen", "config", clusterName, fmt.Sprintf("https://%s:6443", controlplaneIP), "--additional-sans", controlplaneIP, "--config-patch", talosConfigPatch); err != nil {
		return err
	}
	if err := talosCmd(ctx.Dir, "--talosconfig", "talosconfig", "config", "endpoint", controlplaneIP); err != nil {
		return err
	}
	return nil
}

func TalosBootstrap(ctx *Context, serverIP string) error {
	return talosCmd(ctx.Dir, "--talosconfig", "talosconfig", "-n", serverIP, "bootstrap")
}

func TalosKubeconfig(ctx *Context, serverIP string) error {
	return talosCmd(ctx.Dir, "--talosconfig", "talosconfig", "-n", serverIP, "kubeconfig", ".")
}

func TalosCreateControlplane(ctx *Context, network *hcloud.Network, placementGroup *hcloud.PlacementGroup, name string, serverType string) (*hcloud.Server, error) {
	userData, err := ioutil.ReadFile(path.Join(ctx.Dir, "controlplane.yaml"))
	if err != nil {
		return nil, err
	}
	return HcloudCreateServer(ctx, network, placementGroup, name, serverType, "controlplane", string(userData))
}

func TalosCreateWorker(ctx *Context, network *hcloud.Network, placementGroup *hcloud.PlacementGroup, name string, serverType string) (*hcloud.Server, error) {
	userData, err := ioutil.ReadFile(path.Join(ctx.Dir, "worker.yaml"))
	if err != nil {
		return nil, err
	}
	return HcloudCreateServer(ctx, network, placementGroup, name, serverType, "worker", string(userData))
}

func talosCmd(dir string, args ...string) error {
	cmd := exec.Command("talosctl", args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("talos command %s failed: %w\n%s", strings.Join(args, " "), err, output)
	}
	return nil
}
