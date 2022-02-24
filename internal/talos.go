package internal

import (
	_ "embed"
	"fmt"
	"net"
	"os/exec"
	"strings"

	"github.com/airfocusio/hcloud-talos/internal/clients"
	"github.com/airfocusio/hcloud-talos/internal/cluster"
	"github.com/airfocusio/hcloud-talos/internal/utils"
	"github.com/hetznercloud/hcloud-go/hcloud"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8stypes "k8s.io/apimachinery/pkg/types"
)

func TalosGenConfig(cl *cluster.Cluster, network *hcloud.Network, clusterName string, controlplaneIP net.IP, withKubespan bool) (string, error) {
	configPatch, err := utils.RenderTemplate(`
		[
			{
				"op": "add",
				"path": "/machine/kubelet/nodeIP",
				"value": {
					"validSubnets": [
						"{{ .NetworkIPRange }}"
					]
				}
			},
			{
				"op": "add",
				"path": "/cluster/externalCloudProvider",
				"value": {
					"enabled": true,
					"manifests": []
				}
			},
			{
				"op": "add",
				"path": "/machine/systemDiskEncryption",
				"value": {
					"ephemeral": {
						"provider": "luks2",
						"keys": [
							{
								"slot": 0,
								"nodeID": {}
							}
						]
					}
				}
			}
		]
	`, map[string]string{
		"NetworkIPRange": network.IPRange.String(),
	})
	if err != nil {
		return "", err
	}
	args := []string{
		"gen", "config",
		clusterName, fmt.Sprintf("https://%s:6443", controlplaneIP.String()),
		"--additional-sans", controlplaneIP.String(),
		"--config-patch", configPatch,
		"--kubernetes-version", kubernetesVersion,
	}
	if withKubespan {
		args = append(args, "--with-kubespan")
	}
	output1, err := talosCmdRaw(cl.Dir, args...)
	if err != nil {
		return output1, err
	}
	output2, err := talosCmd(cl, "config", "endpoint", controlplaneIP.String())
	if err != nil {
		return output1 + output2, err
	}
	return output1 + output2, nil
}

func TalosBootstrap(cl *cluster.Cluster, serverIP net.IP) (string, error) {
	return talosCmd(cl, "-n", serverIP.String(), "bootstrap")
}

func TalosKubeconfig(cl *cluster.Cluster, serverIP net.IP) (string, error) {
	return talosCmd(cl, "-n", serverIP.String(), "kubeconfig", ".")
}

func TalosReset(cl *cluster.Cluster, serverIP net.IP) (string, error) {
	return talosCmd(cl, "-n", serverIP.String(), "reset", "--system-labels-to-wipe", "STATE", "--system-labels-to-wipe", "EPHEMERAL")
}

func TalosPatchFlannelDaemonSet(cl *cluster.Cluster, jsonPatch string) error {
	kubeClientset, _, err := clients.KubernetesInit(cl)
	if err != nil {
		return err
	}
	_, err = kubeClientset.AppsV1().DaemonSets("kube-system").Patch(*cl.Ctx, "kube-flannel", k8stypes.JSONPatchType, []byte(jsonPatch), k8smetav1.PatchOptions{})
	if err != nil {
		return err
	}
	return nil
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
