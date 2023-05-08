package e2etests

import (
	"testing"

	"github.com/airfocusio/hcloud-talos/internal"
	"github.com/stretchr/testify/assert"
)

func TestAddNode(t *testing.T) {
	_, err := internal.AddNode(&logger, clusterDir, internal.AddNodeOpts{
		ConfigFile:   configFile,
		ServerType:   "cx21",
		NodeName:     "worker-1",
		TalosVersion: talosVersion,
	})
	assert.NoError(t, err)
}
