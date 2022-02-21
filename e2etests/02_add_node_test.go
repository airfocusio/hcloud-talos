package e2etests

import (
	"testing"

	"github.com/airfocusio/hcloud-talos/internal/cmds"
	"github.com/stretchr/testify/assert"
)

func TestAddNode(t *testing.T) {
	cmd := cmds.AddNodeCommand{
		NodeName:       "worker-01",
		NodeServerType: "cx21",
	}
	err := cmd.Run(&logger, clusterDir)
	assert.NoError(t, err)
}
