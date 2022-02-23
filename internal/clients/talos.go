package clients

import (
	_ "embed"
	"fmt"
	"os/exec"
	"strings"

	"github.com/airfocusio/hcloud-talos/internal/cluster"
)

var (
	//go:embed talos_config_patch.json
	talosConfigPatch string
)

func TalosGenConfig(cl *cluster.Cluster, clusterName string, controlplaneIP string) error {
	if err := talosCmd(cl.Dir, "gen", "config", clusterName, fmt.Sprintf("https://%s:6443", controlplaneIP), "--additional-sans", controlplaneIP, "--config-patch", talosConfigPatch); err != nil {
		return err
	}
	if err := talosCmd(cl.Dir, "--talosconfig", "talosconfig", "config", "endpoint", controlplaneIP); err != nil {
		return err
	}
	return nil
}

func TalosBootstrap(cl *cluster.Cluster, serverIP string) error {
	return talosCmd(cl.Dir, "--talosconfig", "talosconfig", "-n", serverIP, "bootstrap")
}

func TalosKubeconfig(cl *cluster.Cluster, serverIP string) error {
	return talosCmd(cl.Dir, "--talosconfig", "talosconfig", "-n", serverIP, "kubeconfig", ".")
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
