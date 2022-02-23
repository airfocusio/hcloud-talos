package e2etests

import (
	"testing"

	"github.com/airfocusio/hcloud-talos/internal"
	"github.com/stretchr/testify/assert"
)

func TestBootstrapCluster(t *testing.T) {
	cmd := internal.BootstrapClusterCommand{
		ClusterName: clusterName,
		ServerType:  "cx21",
		NodeName:    "controlplane-01",
		Location:    "nbg1",
		NetworkZone: "eu-central",
		Token:       hcloudToken,
	}
	err := cmd.Run(&logger, clusterDir)
	assert.NoError(t, err)
}
