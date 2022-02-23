package cluster

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/airfocusio/hcloud-talos/internal/utils"
	"github.com/hetznercloud/hcloud-go/hcloud"
	"gopkg.in/yaml.v3"
)

type Cluster struct {
	Ctx    *context.Context
	Client *hcloud.Client
	Logger *utils.Logger
	Dir    string
	Config Config
}

func (cl *Cluster) Create(logger *utils.Logger, clusterName string, hcloudLocation string, hcloudNetworkZone string, hcloudToken string, force bool) error {
	ctx := context.Background()
	cl.Ctx = &ctx
	cl.Logger = logger

	err := os.MkdirAll(cl.Dir, 0o775)
	if err != nil {
		return err
	}
	files, err := ioutil.ReadDir(cl.Dir)
	if err != nil {
		return err
	} else if len(files) > 0 {
		if !force {
			return fmt.Errorf("directory must be empty")
		} else {
			err := os.RemoveAll(cl.Dir)
			if err != nil {
				return err
			}
			err = os.MkdirAll(cl.Dir, 0o775)
			if err != nil {
				return err
			}
		}
	}

	cl.Config.ClusterName = clusterName
	cl.Config.Hcloud.Location = hcloudLocation
	cl.Config.Hcloud.NetworkZone = hcloudNetworkZone
	cl.Config.Hcloud.Token = hcloudToken

	cl.Client = hcloud.NewClient(hcloud.WithToken(hcloudToken))

	return nil
}

func (cl *Cluster) Load(logger *utils.Logger) error {
	ctx := context.Background()
	cl.Ctx = &ctx
	cl.Logger = logger

	yamlBytes, err := ioutil.ReadFile(path.Join(cl.Dir, configFile))
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(yamlBytes, &cl.Config)
	if err != nil {
		return err
	}

	cl.Client = hcloud.NewClient(hcloud.WithToken(cl.Config.Hcloud.Token))

	return nil
}

func (cl Cluster) Save() error {
	yamlBytes, err := yaml.Marshal(&cl.Config)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(path.Join(cl.Dir, configFile), yamlBytes, 0o664)
	if err != nil {
		return err
	}
	return nil
}
