package e2etests

import (
	"testing"

	"github.com/airfocusio/hcloud-talos/internal"
	"github.com/stretchr/testify/assert"
)

func TestDeleteNode(t *testing.T) {
	t.Skip()

	err := internal.DeleteNode(&logger, clusterDir, internal.DeleteNodeOpts{
		ConfigFile: configFile,
		NodeName:   "worker",
		Force:      true,
	})
	assert.NoError(t, err)
}
