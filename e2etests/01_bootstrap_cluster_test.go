package e2etests

import (
	"testing"

	"github.com/airfocusio/hcloud-talos/internal"
	"github.com/stretchr/testify/assert"
)

func TestBootstrapCluster(t *testing.T) {
	err := internal.BootstrapCluster(&logger, clusterDir, internal.BootstrapClusterOpts{
		ConfigFile:        configFile,
		ClusterName:       clusterName,
		ServerType:        "cx21",
		NodeName:          "controlplane-1",
		Location:          "nbg1",
		NetworkZone:       "eu-central",
		Token:             hcloudToken,
		TalosVersion:      talosVersion,
		KubernetesVersion: kubernetesVersion,
	})
	assert.NoError(t, err)
}
