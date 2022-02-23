package e2etests

import (
	"testing"

	"github.com/airfocusio/hcloud-talos/internal"
	"github.com/stretchr/testify/assert"
)

func TestDeleteNode(t *testing.T) {
	cmd := internal.DeleteNodeCommand{
		NodeName: "worker-01",
		Force:    true,
	}
	err := cmd.Run(&logger, clusterDir)
	assert.NoError(t, err)
}
