package internal

import (
	"flag"
	"fmt"

	"github.com/airfocusio/hcloud-talos/internal/clients"
	"github.com/airfocusio/hcloud-talos/internal/cluster"
	"github.com/airfocusio/hcloud-talos/internal/utils"
)

type AddNodeCommandId struct{}

func (cmdId *AddNodeCommandId) Name() string {
	return "add-node"
}

func (cmdId *AddNodeCommandId) Create() Command {
	return &AddNodeCommand{}
}

type AddNodeCommand struct {
	ServerType   string
	Controlplane bool
	NodeName     string
	PoolName     string
}

var _ Command = (*AddNodeCommand)(nil)

func (cmd *AddNodeCommand) RegisterOpts(flags *flag.FlagSet) {
	flags.BoolVar(&cmd.Controlplane, "controlplane", false, "")
	flags.StringVar(&cmd.ServerType, "server-type", "cx21", "")
	flags.StringVar(&cmd.PoolName, "pool-name", "", "")
	flags.StringVar(&cmd.NodeName, "node-name", "", "")
}

func (cmd *AddNodeCommand) ValidateOpts() error {
	if cmd.NodeName == "" {
		return fmt.Errorf("node name must not be empty")
	}
	if cmd.ServerType == "" {
		return fmt.Errorf("node server type must not be empty")
	}
	return nil
}

func (cmd *AddNodeCommand) Run(logger *utils.Logger, dir string) error {
	cl := &cluster.Cluster{Dir: dir}
	err := cl.Load(logger)
	if err != nil {
		return err
	}
	logger.Info.Printf("Adding node %s/%s (%s)\n", cl.Config.ClusterName, cmd.NodeName, cmd.PoolName)

	network, err := clients.HcloudEnsureNetwork(cl, nodeNetworkTemplate(cl), false)
	if err != nil {
		return err
	}

	placementGroup, err := clients.HcloudEnsurePlacementGroup(cl, nodePlacementGroupTemplate(cl), false)
	if err != nil {
		return err
	}

	var nodeTemplate clients.HcloudServerCreateFromImageOpts
	if cmd.Controlplane {
		nodeTemplate, err = controlplaneNodeTemplate(cl, cmd.ServerType, cmd.NodeName)
		if err != nil {
			return err
		}
	} else {
		nodeTemplate, err = workerNodeTemplate(cl, cmd.ServerType, cmd.PoolName, cmd.NodeName)
		if err != nil {
			return err
		}
	}

	server, err := clients.HcloudCreateServerFromImage(cl, network, placementGroup, nodeTemplate)
	if err != nil {
		return err
	}

	err = clients.KubernetesWaitNodeRunning(cl, server.Name)
	if err != nil {
		return err
	}

	return nil
}
