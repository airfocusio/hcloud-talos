package e2etests

import (
	"testing"

	"github.com/airfocusio/hcloud-talos/internal/cmds"
	"github.com/stretchr/testify/assert"
)

func TestBootstrapCluster(t *testing.T) {
	cmd := cmds.BootstrapClusterCommand{
		ClusterName:    clusterName,
		NodeName:       "controlplane-01",
		NodeServerType: "cx21",
		Location:       "nbg1",
		NetworkZone:    "eu-central",
		Token:          hcloudToken,
	}
	err := cmd.Run(&logger, clusterDir)
	assert.NoError(t, err)
}
