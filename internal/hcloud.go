package internal

import (
	"fmt"
	"net"

	"github.com/airfocusio/hcloud-talos/internal/utils"
	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/hetznercloud/hcloud-go/hcloud/schema"
)

const (
	HcloudLabelPrefix = "hcloudtalos.airfocus.io/"
	HcloudMarkerLabel = HcloudLabelPrefix + "cluster-name"
	HcloudRoleLabel   = HcloudLabelPrefix + "node-role"
)

func HcloudEnsureNetwork(ctx *Context) (*hcloud.Network, error) {
	networkName := ctx.Config.ClusterName + "-nodes"
	network, _, err := ctx.Client.Network.Get(*ctx.Ctx, networkName)
	if err != nil {
		return nil, err
	}
	if network != nil {
		return network, nil
	}

	ctx.Logger.Info.Printf("Creating new network\n")
	_, privateIPRange, _ := net.ParseCIDR("10.0.0.0/16")
	_, privateIPRangeSubnet, _ := net.ParseCIDR("10.0.0.0/24")
	network, _, err = ctx.Client.Network.Create(*ctx.Ctx, hcloud.NetworkCreateOpts{
		Name:    networkName,
		IPRange: privateIPRange,
		Subnets: []hcloud.NetworkSubnet{
			{
				Type:        hcloud.NetworkSubnetTypeCloud,
				IPRange:     privateIPRangeSubnet,
				NetworkZone: hcloud.NetworkZone(ctx.Config.Hcloud.NetworkZone),
			},
		},
		Labels: map[string]string{HcloudMarkerLabel: ctx.Config.ClusterName},
	})
	if err != nil {
		return nil, err
	}

	return network, nil
}

func HcloudEnsurePlacementGroup(ctx *Context) (*hcloud.PlacementGroup, error) {
	placementGroupName := ctx.Config.ClusterName + "-nodes"
	placementGroup, _, err := ctx.Client.PlacementGroup.Get(*ctx.Ctx, placementGroupName)
	if err != nil {
		return nil, err
	}
	if placementGroup != nil {
		return placementGroup, nil
	}

	ctx.Logger.Info.Printf("Creating new placement group\n")
	placementGroupResult, _, err := ctx.Client.PlacementGroup.Create(*ctx.Ctx, hcloud.PlacementGroupCreateOpts{
		Name:   placementGroupName,
		Type:   hcloud.PlacementGroupTypeSpread,
		Labels: map[string]string{HcloudMarkerLabel: ctx.Config.ClusterName},
	})
	if err != nil {
		return nil, err
	}
	placementGroup = placementGroupResult.PlacementGroup

	return placementGroup, nil
}

func HcloudEnsureLoadBalancer(ctx *Context, network *hcloud.Network, name string, services []hcloud.LoadBalancerCreateOptsService, targetRole string) (*hcloud.LoadBalancer, error) {
	loadBalancer, _, err := ctx.Client.LoadBalancer.Get(*ctx.Ctx, name)
	if err != nil {
		return nil, err
	}
	if loadBalancer != nil {
		return loadBalancer, nil
	}

	ctx.Logger.Info.Printf("Creating new load balancer\n")
	loadBalancerResult, _, err := ctx.Client.LoadBalancer.Create(*ctx.Ctx, hcloud.LoadBalancerCreateOpts{
		Name: name,
		LoadBalancerType: &hcloud.LoadBalancerType{
			Name: "lb11",
		},
		Location: &hcloud.Location{
			Name: ctx.Config.Hcloud.Location,
		},
		Network:  network,
		Services: services,
		Labels:   map[string]string{HcloudMarkerLabel: ctx.Config.ClusterName},
	})
	if err != nil {
		return nil, err
	}
	loadBalancer = loadBalancerResult.LoadBalancer

	err = utils.Retry(ctx.Logger, func() error {
		loadBalancer, _, err = ctx.Client.LoadBalancer.GetByID(*ctx.Ctx, loadBalancer.ID)
		if err != nil {
			return err
		}
		if loadBalancer.PublicNet.IPv4.IP == nil {
			return fmt.Errorf("load balancer does not yet have a public IP")
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	err = utils.Retry(ctx.Logger, func() error {
		loadBalancer, _, err = ctx.Client.LoadBalancer.GetByID(*ctx.Ctx, loadBalancer.ID)
		if err != nil {
			return err
		}
		if len(loadBalancer.PrivateNet) == 0 || loadBalancer.PrivateNet[0].IP == nil {
			return fmt.Errorf("load balancer does not yet have a private IP")
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	err = utils.Retry(ctx.Logger, func() error {
		usePrivateIP := true
		_, _, err := ctx.Client.LoadBalancer.AddLabelSelectorTarget(*ctx.Ctx, loadBalancer, hcloud.LoadBalancerAddLabelSelectorTargetOpts{
			Selector:     HcloudMarkerLabel + "=" + ctx.Config.ClusterName + "," + HcloudRoleLabel + "=" + targetRole,
			UsePrivateIP: &usePrivateIP,
		})
		return err
	})
	if err != nil {
		return nil, err
	}

	return loadBalancer, nil
}

func HcloudEnsureFirewall(ctx *Context) (*hcloud.Firewall, error) {
	firewallName := ctx.Config.ClusterName + "-nodes"
	firewall, _, err := ctx.Client.Firewall.Get(*ctx.Ctx, firewallName)
	if err != nil {
		return nil, err
	}
	if firewall != nil {
		return firewall, nil
	}

	ctx.Logger.Info.Printf("Creating new firewall\n")
	_, publicIPRange4, _ := net.ParseCIDR("0.0.0.0/0")
	_, publicIPRange6, _ := net.ParseCIDR("::/0")
	_, privateIPRange, _ := net.ParseCIDR("10.0.0.0/16")
	anyPort := "any"
	firewallResult, _, err := ctx.Client.Firewall.Create(*ctx.Ctx, hcloud.FirewallCreateOpts{
		Name: firewallName,
		Rules: []hcloud.FirewallRule{
			{
				Direction: hcloud.FirewallRuleDirectionIn,
				Protocol:  hcloud.FirewallRuleProtocolTCP,
				SourceIPs: []net.IPNet{*privateIPRange},
				Port:      &anyPort,
			},
			{
				Direction: hcloud.FirewallRuleDirectionIn,
				Protocol:  hcloud.FirewallRuleProtocolUDP,
				SourceIPs: []net.IPNet{*privateIPRange},
				Port:      &anyPort,
			},
			{
				Direction: hcloud.FirewallRuleDirectionIn,
				Protocol:  hcloud.FirewallRuleProtocolICMP,
				SourceIPs: []net.IPNet{*privateIPRange},
			},
			{
				Direction: hcloud.FirewallRuleDirectionIn,
				Protocol:  hcloud.FirewallRuleProtocolICMP,
				SourceIPs: []net.IPNet{*publicIPRange4, *publicIPRange6},
			},
		},
		ApplyTo: []hcloud.FirewallResource{
			{
				Type: hcloud.FirewallResourceTypeLabelSelector,
				LabelSelector: &hcloud.FirewallResourceLabelSelector{
					Selector: HcloudMarkerLabel + "=" + ctx.Config.ClusterName + "," + HcloudRoleLabel,
				},
			},
		},
		Labels: map[string]string{HcloudMarkerLabel: ctx.Config.ClusterName},
	})
	if err != nil {
		return nil, err
	}
	firewall = firewallResult.Firewall

	return firewall, nil
}

func HcloudCreateServer(ctx *Context, network *hcloud.Network, placementGroup *hcloud.PlacementGroup, name string, serverType string, role string, userData string) (*hcloud.Server, error) {
	nodeName := ctx.Config.ClusterName + "-" + name

	ctx.Logger.Info.Printf("Creating new server\n")
	ctx.Logger.Debug.Printf("Generating temporary SSH key\n")
	sshKeyPrivate := utils.SSHKeyPrivate{}
	if err := sshKeyPrivate.Generate(); err != nil {
		return nil, err
	}
	sshKeyPublic, err := sshKeyPrivate.StorePublic()
	if err != nil {
		return nil, err
	}
	ctx.Logger.Debug.Printf("Registering temporary SSH key\n")
	sshKey, _, err := ctx.Client.SSHKey.Create(*ctx.Ctx, hcloud.SSHKeyCreateOpts{
		Name:      nodeName + "-init-" + utils.RandString(8),
		PublicKey: sshKeyPublic,
		Labels:    map[string]string{HcloudMarkerLabel: ctx.Config.ClusterName},
	})
	if err != nil {
		return nil, err
	}
	defer func() {
		ctx.Logger.Debug.Printf("Removing temporary SSH key\n")
		ctx.Client.SSHKey.Delete(*ctx.Ctx, sshKey)
	}()

	initImage := "debian-11"
	startAfterCreate := false
	serverRespone, _, err := ctx.Client.Server.Create(*ctx.Ctx, hcloud.ServerCreateOpts{
		Name: nodeName,
		ServerType: hcloud.ServerTypeFromSchema(schema.ServerType{
			Name: serverType,
		}),
		Image: hcloud.ImageFromSchema(schema.Image{
			Name: &initImage,
		}),
		PlacementGroup: placementGroup,
		Location: &hcloud.Location{
			Name: ctx.Config.Hcloud.Location,
		},
		Networks: []*hcloud.Network{
			network,
		},
		StartAfterCreate: &startAfterCreate,
		UserData:         userData,
		SSHKeys:          []*hcloud.SSHKey{sshKey}, // not needed, but without Hetzner sends out server creation email
		Labels:           map[string]string{HcloudMarkerLabel: ctx.Config.ClusterName},
	})
	if err != nil {
		return nil, err
	}

	ctx.Logger.Debug.Printf("Waiting for server to have private IP\n")
	server := serverRespone.Server
	err = utils.Retry(ctx.Logger, func() error {
		server, _, err = ctx.Client.Server.GetByID(*ctx.Ctx, server.ID)
		if err != nil {
			return err
		}
		if len(server.PrivateNet) == 0 || server.PrivateNet[0].IP == nil {
			return fmt.Errorf("server does not yet have a private IP")
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	ctx.Logger.Debug.Printf("Starting server in rescue mode\n")
	err = utils.Retry(ctx.Logger, func() error {
		_, _, err = ctx.Client.Server.EnableRescue(*ctx.Ctx, server, hcloud.ServerEnableRescueOpts{
			Type:    hcloud.ServerRescueTypeLinux64,
			SSHKeys: []*hcloud.SSHKey{sshKey},
		})
		return err
	})
	if err != nil {
		return nil, err
	}
	err = utils.Retry(ctx.Logger, func() error {
		_, _, err = ctx.Client.Server.Poweron(*ctx.Ctx, server)
		return err
	})
	if err != nil {
		return nil, err
	}
	err = utils.RetrySlow(ctx.Logger, func() error {
		_, err := sshKeyPrivate.Execute(server.PublicNet.IPv4.IP.String(), 22, "true")
		return err
	})
	if err != nil {
		return nil, err
	}

	ctx.Logger.Debug.Printf("Applying talos image\n")
	err = utils.Retry(ctx.Logger, func() error {
		_, err := sshKeyPrivate.Execute(server.PublicNet.IPv4.IP.String(), 22, `
			cd /tmp
			wget -O /tmp/talos.raw.xz https://github.com/talos-systems/talos/releases/download/v0.14.2/hcloud-amd64.raw.xz
			xz -d -c /tmp/talos.raw.xz | dd of=/dev/sda && sync
		`)
		return err
	})
	if err != nil {
		return nil, err
	}

	ctx.Logger.Debug.Printf("Shutting down server\n")
	err = utils.Retry(ctx.Logger, func() error {
		_, _, err = ctx.Client.Server.Shutdown(*ctx.Ctx, server)
		return err
	})
	if err != nil {
		return nil, err
	}
	err = utils.Retry(ctx.Logger, func() error {
		server, _, err := ctx.Client.Server.GetByID(*ctx.Ctx, server.ID)
		if err != nil {
			return err
		}
		if server.Status != hcloud.ServerStatusOff {
			return fmt.Errorf("server is not yet turned off")
		}
		return nil
	})

	ctx.Logger.Debug.Printf("Configuring server labels\n")
	err = utils.Retry(ctx.Logger, func() error {
		server, _, err = ctx.Client.Server.Update(*ctx.Ctx, server, hcloud.ServerUpdateOpts{
			Labels: map[string]string{
				HcloudMarkerLabel: ctx.Config.ClusterName,
				HcloudRoleLabel:   role,
			},
		})
		return err
	})
	if err != nil {
		return nil, err
	}

	ctx.Logger.Debug.Printf("Starting server\n")
	err = utils.Retry(ctx.Logger, func() error {
		_, _, err = ctx.Client.Server.Poweron(*ctx.Ctx, server)
		return err
	})
	if err != nil {
		return nil, err
	}

	return server, nil
}
