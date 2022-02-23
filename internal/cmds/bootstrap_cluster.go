package cmds

import (
	"flag"
	"fmt"
	"os"

	"github.com/airfocusio/hcloud-talos/internal"
	"github.com/airfocusio/hcloud-talos/internal/utils"
	"github.com/hetznercloud/hcloud-go/hcloud"
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
	NodeServerType        string
	Location              string
	NetworkZone           string
	Token                 string
	SkipIngress           bool
	SkipFirewall          bool
	Force                 bool
	ApplyManifestsCommand ApplyManifestsCommand
}

func (cmd *BootstrapClusterCommand) RegisterOpts(flags *flag.FlagSet) {
	cmd.Token = os.Getenv("HCLOUD_TOKEN")
	flags.StringVar(&cmd.ClusterName, "cluster-name", "", "")
	flags.StringVar(&cmd.NodeName, "node-name", "", "")
	flags.StringVar(&cmd.NodeServerType, "node-server-type", "cx21", "")
	flags.StringVar(&cmd.Location, "location", "nbg1", "")
	flags.StringVar(&cmd.NetworkZone, "network-zone", "eu-central", "")
	flags.BoolVar(&cmd.SkipIngress, "skip-ingress", false, "")
	flags.BoolVar(&cmd.SkipFirewall, "skip-firewall", false, "")
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
	if cmd.NodeServerType == "" {
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
	ctx := &internal.Context{Dir: dir}
	err := ctx.Create(logger, cmd.ClusterName, cmd.Location, cmd.NetworkZone, cmd.Token, cmd.Force)
	if err != nil {
		return err
	}
	err = ctx.Save()
	if err != nil {
		return err
	}

	network, err := internal.HcloudEnsureNetwork(ctx)
	if err != nil {
		return err
	}

	placementGroup, err := internal.HcloudEnsurePlacementGroup(ctx)
	if err != nil {
		return err
	}

	kubernetesApiServerPort := 6443
	talosApiServerPort := 50000
	controlplaneLoadBalancerName := ctx.Config.ClusterName + "-controlplane"
	controlplaneLoadBalancerServices := []hcloud.LoadBalancerCreateOptsService{
		{
			ListenPort:      &kubernetesApiServerPort,
			DestinationPort: &kubernetesApiServerPort,
			Protocol:        hcloud.LoadBalancerServiceProtocolTCP,
		},
		{
			ListenPort:      &talosApiServerPort,
			DestinationPort: &talosApiServerPort,
			Protocol:        hcloud.LoadBalancerServiceProtocolTCP,
		},
	}
	controlplaneLoadBalancer, err := internal.HcloudEnsureLoadBalancer(ctx, network, controlplaneLoadBalancerName, controlplaneLoadBalancerServices, "controlplane")
	if err != nil {
		return err
	}

	if !cmd.SkipIngress {
		ingressLoadBalancerName := ctx.Config.ClusterName + "-ingress"
		ingressHttpPortIn := 80
		ingressHttpPortOut := 30080
		ingressHttpsPortIn := 443
		ingressHttpsPortOut := 30433
		ingressLoadBalancerProxyProtocol := true
		ingressLoadBalancerServices := []hcloud.LoadBalancerCreateOptsService{
			{
				ListenPort:      &ingressHttpPortIn,
				DestinationPort: &ingressHttpPortOut,
				Protocol:        hcloud.LoadBalancerServiceProtocolTCP,
				Proxyprotocol:   &ingressLoadBalancerProxyProtocol,
			},
			{
				ListenPort:      &ingressHttpsPortIn,
				DestinationPort: &ingressHttpsPortOut,
				Protocol:        hcloud.LoadBalancerServiceProtocolTCP,
				Proxyprotocol:   &ingressLoadBalancerProxyProtocol,
			},
		}
		_, err = internal.HcloudEnsureLoadBalancer(ctx, network, ingressLoadBalancerName, ingressLoadBalancerServices, "worker")
		if err != nil {
			return err
		}
	}

	if !cmd.SkipFirewall {
		_, err = internal.HcloudEnsureFirewall(ctx)
		if err != nil {
			return err
		}
	}

	err = internal.TalosGenConfig(ctx, cmd.ClusterName, controlplaneLoadBalancer.PublicNet.IPv4.IP.String())
	if err != nil {
		return err
	}

	controlplaneServer, err := internal.TalosCreateControlplane(ctx, network, placementGroup, cmd.NodeName, cmd.NodeServerType)
	if err != nil {
		return err
	}
	controlplaneServerPrivateIP := controlplaneServer.PrivateNet[0].IP

	err = utils.RetrySlow(logger, func() error {
		return internal.TalosBootstrap(ctx, controlplaneServerPrivateIP.String())
	})
	if err != nil {
		return err
	}

	err = utils.Retry(logger, func() error {
		return internal.TalosKubeconfig(ctx, controlplaneServerPrivateIP.String())
	})
	if err != nil {
		return err
	}

	err = utils.RetrySlow(logger, func() error {
		nodes, err := internal.KubernetesListNodes(ctx)
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
