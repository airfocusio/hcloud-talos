package internal

import (
	"fmt"

	"github.com/airfocusio/hcloud-talos/internal/clients"
	"github.com/airfocusio/hcloud-talos/internal/cluster"
	"github.com/airfocusio/hcloud-talos/internal/utils"
	"github.com/hetznercloud/hcloud-go/hcloud"
)

type AddNodeOpts struct {
	ConfigFile   string
	ServerType   string
	Controlplane bool
	NodeName     string
	PoolName     string
	TalosVersion string
}

func AddNode(logger *utils.Logger, dir string, opts AddNodeOpts) (*hcloud.Server, error) {
	cl := &cluster.Cluster{Dir: dir}
	err := cl.Load(opts.ConfigFile, logger)
	if err != nil {
		return nil, err
	}
	logger.Info.Printf("Adding node %s/%s (%s)\n", cl.Config.ClusterName, opts.NodeName, opts.PoolName)
	if opts.NodeName == "" {
		return nil, fmt.Errorf("node name must not be empty")
	}
	if opts.ServerType == "" {
		return nil, fmt.Errorf("node server type must not be empty")
	}
	if opts.TalosVersion == "" {
		return nil, fmt.Errorf("talos version must not be empty")
	}

	network, err := clients.HcloudEnsureNetwork(cl, nodeNetworkTemplate(cl), false)
	if err != nil {
		return nil, err
	}

	var placementGroup *hcloud.PlacementGroup
	if opts.Controlplane {
		controlplanePlacementGroup, err := clients.HcloudEnsurePlacementGroup(cl, controlplanePlacementGroupTemplate(cl), false)
		if err != nil {
			return nil, err
		}
		placementGroup = controlplanePlacementGroup
	}

	var nodeTemplate clients.HcloudServerCreateFromImageOpts
	if opts.Controlplane {
		nodeTemplate, err = controlplaneNodeTemplate(cl, opts.ServerType, opts.NodeName, opts.TalosVersion)
		if err != nil {
			return nil, err
		}
	} else {
		nodeTemplate, err = workerNodeTemplate(cl, opts.ServerType, opts.PoolName, opts.NodeName, opts.TalosVersion)
		if err != nil {
			return nil, err
		}
	}

	server, err := clients.HcloudCreateServerFromImage(cl, network, placementGroup, nodeTemplate)
	if err != nil {
		return nil, err
	}

	err = clients.KubernetesWaitNodeRegistered(cl, server.Name)
	if err != nil {
		return nil, err
	}

	return server, nil
}
