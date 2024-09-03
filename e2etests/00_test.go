package e2etests

import (
	"fmt"
	"os"
	"path"
	"runtime"
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

var talosVersion string
var talosctlUrl string
var talosctlBin string
var kubernetesVersion string

func TestMain(t *testing.M) {
	setup()
	exitCode := t.Run()
	cleanup()
	os.Exit(exitCode)
}

func setup() {
	talosVersion = os.Getenv("TALOS_VERSION")
	if talosVersion == "" {
		talosVersion = "1.7.6"
	}
	talosctlUrl = fmt.Sprintf("https://github.com/siderolabs/talos/releases/download/v%s/talosctl-%s-%s", talosVersion, runtime.GOOS, runtime.GOARCH)
	talosctlBin = path.Join(os.TempDir(), fmt.Sprintf("talosctl-%s", talosVersion))
	PrepareBinaries(talosctlUrl, &RawBinariesUnpack{Name: talosctlBin})
	internal.TalosctlBin = talosctlBin

	version, err := internal.TalosClientVersion()
	if err != nil {
		fmt.Printf("unable to retrieve talos client version: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("talos client version %s\n", version)

	kubernetesVersion = os.Getenv("KUBERNETES_VERSION")
	if kubernetesVersion == "" {
		kubernetesVersion = "1.30.4"
	}

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
