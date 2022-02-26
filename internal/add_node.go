package internal

import (
	"fmt"

	"github.com/airfocusio/hcloud-talos/internal/clients"
	"github.com/airfocusio/hcloud-talos/internal/cluster"
	"github.com/airfocusio/hcloud-talos/internal/utils"
	"github.com/hetznercloud/hcloud-go/hcloud"
)

type AddNodeOpts struct {
	ServerType   string
	Controlplane bool
	NodeName     string
	PoolName     string
}

func AddNode(logger *utils.Logger, dir string, opts AddNodeOpts) (*hcloud.Server, error) {
	cl := &cluster.Cluster{Dir: dir}
	err := cl.Load(logger)
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

	network, err := clients.HcloudEnsureNetwork(cl, nodeNetworkTemplate(cl), false)
	if err != nil {
		return nil, err
	}

	placementGroup, err := clients.HcloudEnsurePlacementGroup(cl, nodePlacementGroupTemplate(cl), false)
	if err != nil {
		return nil, err
	}

	var nodeTemplate clients.HcloudServerCreateFromImageOpts
	if opts.Controlplane {
		nodeTemplate, err = controlplaneNodeTemplate(cl, opts.ServerType, opts.NodeName)
		if err != nil {
			return nil, err
		}
	} else {
		nodeTemplate, err = workerNodeTemplate(cl, opts.ServerType, opts.PoolName, opts.NodeName)
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
