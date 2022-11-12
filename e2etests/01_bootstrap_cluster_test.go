package e2etests

import (
	"os"
	"testing"

	"github.com/airfocusio/hcloud-talos/internal"
	"github.com/stretchr/testify/assert"
)

func TestBootstrapCluster(t *testing.T) {
	talosVersion := os.Getenv("TALOS_VERSION")
	if talosVersion == "" {
		talosVersion = "1.0.5"
	}
	kubernetesVersion := os.Getenv("KUBERNETES_VERSION")
	if kubernetesVersion == "" {
		kubernetesVersion = "1.24.7"
	}

	err := internal.BootstrapCluster(&logger, clusterDir, internal.BootstrapClusterOpts{
		ConfigFile:        configFile,
		ClusterName:       clusterName,
		ServerType:        "cx21",
		NodeName:          "controlplane-01",
		Location:          "nbg1",
		NetworkZone:       "eu-central",
		Token:             hcloudToken,
		TalosVersion:      talosVersion,
		KubernetesVersion: kubernetesVersion,
	})
	assert.NoError(t, err)
}
