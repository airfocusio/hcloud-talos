package e2etests

import (
	_ "embed"
	"testing"

	"github.com/airfocusio/hcloud-talos/internal/clients"
	"github.com/airfocusio/hcloud-talos/internal/cluster"
	"github.com/airfocusio/hcloud-talos/internal/utils"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	//go:embed 03_volumes_test_manifest.yaml
	manifest string
)

func TestVolumes(t *testing.T) {
	cl := &cluster.Cluster{Dir: clusterDir}
	err := cl.Load(&logger)
	assert.NoError(t, err)

	manifests, err := utils.YamlSplitMany([]byte(manifest))
	assert.NoError(t, err)
	for _, manifest := range manifests {
		err = utils.Retry(cl.Logger, func() error {
			return clients.KubernetesCreateFromManifest(cl, string(manifest))
		})
		assert.NoError(t, err)
	}

	err = clients.KubernetesWaitPodRunning(cl, "test", "test-0")
	assert.NoError(t, err)

	clientset, _, err := clients.KubernetesInit(cl)
	assert.NoError(t, err)
	err = clientset.CoreV1().Namespaces().Delete(*cl.Ctx, "test", v1.DeleteOptions{})
	assert.NoError(t, err)
}
