package e2etests

import (
	"testing"

	"github.com/airfocusio/hcloud-talos/internal"
	"github.com/stretchr/testify/assert"
)

func TestAddNode(t *testing.T) {
	cmd := internal.AddNodeCommand{
		ServerType: "cx21",
		NodeName:   "worker-01",
	}
	err := cmd.Run(&logger, clusterDir)
	assert.NoError(t, err)
}
