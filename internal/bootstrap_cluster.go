package internal

import (
	"flag"
	"fmt"
	"os"

	"github.com/airfocusio/hcloud-talos/internal/clients"
	"github.com/airfocusio/hcloud-talos/internal/cluster"
	"github.com/airfocusio/hcloud-talos/internal/utils"
)

type BootstrapClusterCommandId struct{}

func (cmdId *BootstrapClusterCommandId) Name() string {
	return "bootstrap-cluster"
}

func (cmdId *BootstrapClusterCommandId) Create() Command {
	return &BootstrapClusterCommand{}
}

type BootstrapClusterCommand struct {
	ClusterName           string
	NodeName              string
	ServerType            string
	Location              string
	NetworkZone           string
	Token                 string
	NoFirewall            bool
	NoKubespan            bool
	Force                 bool
	ApplyManifestsCommand ApplyManifestsCommand
}

func (cmd *BootstrapClusterCommand) RegisterOpts(flags *flag.FlagSet) {
	cmd.Token = os.Getenv("HCLOUD_TOKEN")
	flags.StringVar(&cmd.ClusterName, "cluster-name", "", "")
	flags.StringVar(&cmd.NodeName, "node-name", "", "")
	flags.StringVar(&cmd.ServerType, "server-type", "cx21", "")
	flags.StringVar(&cmd.Location, "location", "nbg1", "")
	flags.StringVar(&cmd.NetworkZone, "network-zone", "eu-central", "")
	flags.BoolVar(&cmd.NoFirewall, "no-firewall", false, "")
	flags.BoolVar(&cmd.NoKubespan, "no-kubespan", false, "")
	flags.BoolVar(&cmd.Force, "force", false, "")
	cmd.ApplyManifestsCommand.RegisterOpts(flags)
}

func (cmd *BootstrapClusterCommand) ValidateOpts() error {
	if cmd.ClusterName == "" {
		return fmt.Errorf("cluster name must not be empty")
	}
	if cmd.NodeName == "" {
		return fmt.Errorf("node name must not be empty")
	}
	if cmd.ServerType == "" {
		return fmt.Errorf("node server type must not be empty")
	}
	if cmd.Location == "" {
		return fmt.Errorf("location must not be empty")
	}
	if cmd.NetworkZone == "" {
		return fmt.Errorf("network zone must not be empty")
	}
	if cmd.Token == "" {
		return fmt.Errorf("token must not be empty")
	}
	if err := cmd.ApplyManifestsCommand.ValidateOpts(); err != nil {
		return err
	}
	return nil
}

func (cmd *BootstrapClusterCommand) Run(logger *utils.Logger, dir string) error {
	cl := &cluster.Cluster{Dir: dir}
	err := cl.Create(logger, cmd.ClusterName, cmd.Location, cmd.NetworkZone, cmd.Token, cmd.Force)
	if err != nil {
		return err
	}
	err = cl.Save()
	if err != nil {
		return err
	}
	logger.Info.Printf("Bootstrapping cluster %s\n", cl.Config.ClusterName)

	network, err := clients.HcloudEnsureNetwork(cl, nodeNetworkTemplate(cl), true)
	if err != nil {
		return err
	}

	placementGroup, err := clients.HcloudEnsurePlacementGroup(cl, nodePlacementGroupTemplate(cl), true)
	if err != nil {
		return err
	}

	controlplaneLoadBalancer, err := clients.HcloudEnsureLoadBalancer(cl, network, controlPlaneLoadBalanacerTemplate(cl, network), true)
	if err != nil {
		return err
	}

	if !cmd.NoFirewall {
		_, err = clients.HcloudEnsureFirewall(cl, nodeFirewallTemplate(cl, network), true)
		if err != nil {
			return err
		}
	}

	_, err = clients.TalosGenConfig(cl, cmd.ClusterName, controlplaneLoadBalancer.PublicNet.IPv4.IP.String(), !cmd.NoKubespan)
	if err != nil {
		return err
	}

	controlPlaneNodeTemplate, err := controlPlaneNodeTemplate(cl, cmd.ServerType, cmd.NodeName)
	if err != nil {
		return err
	}
	controlplaneServer, err := clients.HcloudCreateServerFromImage(cl, network, placementGroup, controlPlaneNodeTemplate)
	if err != nil {
		return err
	}
	controlplaneServerPrivateIP := controlplaneServer.PrivateNet[0].IP

	err = utils.RetrySlow(logger, func() error {
		_, err := clients.TalosBootstrap(cl, controlplaneServerPrivateIP.String())
		return err
	})
	if err != nil {
		return err
	}

	err = utils.Retry(logger, func() error {
		_, err := clients.TalosKubeconfig(cl, controlplaneServerPrivateIP.String())
		return err
	})
	if err != nil {
		return err
	}

	err = utils.RetrySlow(logger, func() error {
		nodes, err := clients.KubernetesListNodes(cl)
		if err != nil {
			return err
		}
		for _, node := range nodes.Items {
			if node.Name == controlplaneServer.Name {
				return nil
			}
		}
		return fmt.Errorf("node not yet available")
	})
	if err != nil {
		return err
	}

	applyManifestsCommand := ApplyManifestsCommand{}
	err = applyManifestsCommand.Run(logger, dir)
	if err != nil {
		return err
	}

	return nil
}
