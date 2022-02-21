package internal

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

type Context struct {
	Ctx    *context.Context
	Client *hcloud.Client
	Logger *utils.Logger
	Dir    string
	Config Config
}

type Config struct {
	ClusterName string       `yaml:"clusterName"`
	Hcloud      ConfigHcloud `yaml:"hcloud"`
}

type ConfigHcloud struct {
	Location    string `yaml:"location"`
	NetworkZone string `yaml:"networkZone"`
	Token       string `yaml:"token"`
}

const configFile = "hcloudtalosconfig"

func (c *Context) Create(logger *utils.Logger, clusterName string, hcloudLocation string, hcloudNetworkZone string, hcloudToken string, force bool) error {
	ctx := context.Background()
	c.Ctx = &ctx
	c.Logger = logger

	err := os.MkdirAll(c.Dir, 0o775)
	if err != nil {
		return err
	}
	files, err := ioutil.ReadDir(c.Dir)
	if err != nil {
		return err
	} else if len(files) > 0 {
		if !force {
			return fmt.Errorf("directory must be empty")
		} else {
			err := os.RemoveAll(c.Dir)
			if err != nil {
				return err
			}
			err = os.MkdirAll(c.Dir, 0o775)
			if err != nil {
				return err
			}
		}
	}

	c.Config.ClusterName = clusterName
	c.Config.Hcloud.Location = hcloudLocation
	c.Config.Hcloud.NetworkZone = hcloudNetworkZone
	c.Config.Hcloud.Token = hcloudToken

	c.Client = hcloud.NewClient(hcloud.WithToken(hcloudToken))

	return nil
}

func (c *Context) Load(logger *utils.Logger) error {
	ctx := context.Background()
	c.Ctx = &ctx
	c.Logger = logger

	yamlBytes, err := ioutil.ReadFile(path.Join(c.Dir, configFile))
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(yamlBytes, &c.Config)
	if err != nil {
		return err
	}

	c.Client = hcloud.NewClient(hcloud.WithToken(c.Config.Hcloud.Token))

	return nil
}

func (c Context) Save() error {
	yamlBytes, err := yaml.Marshal(&c.Config)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(path.Join(c.Dir, configFile), yamlBytes, 0o664)
	if err != nil {
		return err
	}
	return nil
}
