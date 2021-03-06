package e2etests

import (
	"fmt"
	"os"
	"testing"

	"github.com/airfocusio/hcloud-talos/internal"
	"github.com/airfocusio/hcloud-talos/internal/utils"
)

var (
	clusterName = ""
	clusterDir  = ""
	hcloudToken = ""
	logger      = utils.NewLogger(true)
)

const configFile = "hcloud-talos.yaml"

func TestMain(t *testing.M) {
	setup()
	exitCode := t.Run()
	cleanup()
	os.Exit(exitCode)
}

func setup() {
	fmt.Printf("setup\n")

	hcloudToken = os.Getenv("HCLOUD_TOKEN")
	if hcloudToken == "" {
		fmt.Printf("HCLOUD_TOKEN environment variable is missing\n")
		os.Exit(0)
	}

	clusterName = "test-cluster-" + utils.RandString(8)
	fmt.Printf("cluster name %s\n", clusterName)

	clusterDirOut, err := os.MkdirTemp(os.TempDir(), "")
	if err != nil {
		fmt.Printf("unable to create temporary dir: %v\n", err)
		os.Exit(1)
	}
	clusterDir = clusterDirOut
	fmt.Printf("cluster dir %s\n", clusterDir)
}

func cleanup() {
	fmt.Printf("cleanup\n")

	err := internal.DestroyCluster(&logger, clusterDir, internal.DestroyClusterOpts{
		ConfigFile: configFile,
		Force:      true,
	})
	if err != nil {
		fmt.Printf("unable to destroy cluster: %v\n", err)
	}
}
