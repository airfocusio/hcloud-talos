package internal

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

func TalosGenConfig(cl *cluster.Cluster, clusterName string, controlplaneIP string, withKubespan bool) (string, error) {
	args := []string{
		"gen", "config",
		clusterName, fmt.Sprintf("https://%s:6443", controlplaneIP),
		"--additional-sans", controlplaneIP,
		"--config-patch", talosConfigPatch,
		"--kubernetes-version", kubernetesVersion,
	}
	if withKubespan {
		args = append(args, "--with-kubespan")
	}
	output1, err := talosCmdRaw(cl.Dir, args...)
	if err != nil {
		return output1, err
	}
	output2, err := talosCmd(cl, "config", "endpoint", controlplaneIP)
	if err != nil {
		return output1 + output2, err
	}
	return output1 + output2, nil
}

func TalosBootstrap(cl *cluster.Cluster, serverIP string) (string, error) {
	return talosCmd(cl, "-n", serverIP, "bootstrap")
}

func TalosKubeconfig(cl *cluster.Cluster, serverIP string) (string, error) {
	return talosCmd(cl, "-n", serverIP, "kubeconfig", ".")
}

func TalosReset(cl *cluster.Cluster, serverIP string) (string, error) {
	return talosCmd(cl, "-n", serverIP, "reset", "--system-labels-to-wipe", "STATE", "--system-labels-to-wipe", "EPHEMERAL")
}

func talosCmd(cl *cluster.Cluster, args ...string) (string, error) {
	fullArgs := append([]string{"--talosconfig", "talosconfig"}, args...)
	return talosCmdRaw(cl.Dir, fullArgs...)
}

func talosCmdRaw(dir string, args ...string) (string, error) {
	cmd := exec.Command("talosctl", args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("talos command %s failed: %w\n%s", strings.Join(args, " "), err, output)
	}
	return string(output), nil
}
