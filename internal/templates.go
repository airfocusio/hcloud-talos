package internal

import (
	"io/ioutil"
	"net"
	"path"

	"github.com/airfocusio/hcloud-talos/internal/clients"
	"github.com/airfocusio/hcloud-talos/internal/cluster"
	"github.com/hetznercloud/hcloud-go/hcloud"
)

const (
	labelPrefix  = "hct.airfocus.io/"
	clusterLabel = labelPrefix + "cluster"
	roleLabel    = labelPrefix + "role"
	poolLabel    = labelPrefix + "pool"
)

func nodeNetworkTemplate(cl *cluster.Cluster) hcloud.NetworkCreateOpts {
	_, privateIPRange, _ := net.ParseCIDR("10.0.0.0/16")
	_, privateIPRangeSubnet, _ := net.ParseCIDR("10.0.0.0/24")
	return hcloud.NetworkCreateOpts{
		Name:    cl.Config.ClusterName + "-nodes",
		IPRange: privateIPRange,
		Subnets: []hcloud.NetworkSubnet{
			{
				Type:        hcloud.NetworkSubnetTypeCloud,
				IPRange:     privateIPRangeSubnet,
				NetworkZone: hcloud.NetworkZone(cl.Config.Hcloud.NetworkZone),
			},
		},
		Labels: map[string]string{clusterLabel: cl.Config.ClusterName},
	}
}

func nodePlacementGroupTemplate(cl *cluster.Cluster) hcloud.PlacementGroupCreateOpts {
	return hcloud.PlacementGroupCreateOpts{
		Name:   cl.Config.ClusterName + "-nodes",
		Type:   hcloud.PlacementGroupTypeSpread,
		Labels: map[string]string{clusterLabel: cl.Config.ClusterName},
	}
}

func controlPlaneLoadBalanacerTemplate(cl *cluster.Cluster, network *hcloud.Network) hcloud.LoadBalancerCreateOpts {
	kubernetesApiServerPort := 6443
	talosApiServerPort := 50000
	usePrivateIP := true
	return hcloud.LoadBalancerCreateOpts{
		Name: cl.Config.ClusterName + "-controlplane",
		LoadBalancerType: &hcloud.LoadBalancerType{
			Name: "lb11",
		},
		Location: &hcloud.Location{
			Name: cl.Config.Hcloud.Location,
		},
		Network: network,
		Services: []hcloud.LoadBalancerCreateOptsService{
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
		},
		Targets: []hcloud.LoadBalancerCreateOptsTarget{
			{
				Type: hcloud.LoadBalancerTargetTypeLabelSelector,
				LabelSelector: hcloud.LoadBalancerCreateOptsTargetLabelSelector{
					Selector: clusterLabel + "=" + cl.Config.ClusterName + "," + roleLabel + "=controlplane",
				},
				UsePrivateIP: &usePrivateIP,
			},
		},
		Labels: map[string]string{clusterLabel: cl.Config.ClusterName},
	}
}

func nodeFirewallTemplate(cl *cluster.Cluster, network *hcloud.Network) hcloud.FirewallCreateOpts {
	_, publicIPRange4, _ := net.ParseCIDR("0.0.0.0/0")
	_, publicIPRange6, _ := net.ParseCIDR("::/0")
	privateIPRange := network.IPRange
	anyPort := "any"
	return hcloud.FirewallCreateOpts{
		Name: cl.Config.ClusterName + "-nodes",
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
					Selector: clusterLabel + "=" + cl.Config.ClusterName + "," + roleLabel,
				},
			},
		},
		Labels: map[string]string{clusterLabel: cl.Config.ClusterName},
	}
}

func nodeName(cl *cluster.Cluster, name string) string {
	return cl.Config.ClusterName + "-" + name
}

func controlPlaneNodeTemplate(cl *cluster.Cluster, serverType string, name string) (clients.HcloudServerCreateFromImageOpts, error) {
	userData, err := ioutil.ReadFile(path.Join(cl.Dir, "controlplane.yaml"))
	if err != nil {
		return clients.HcloudServerCreateFromImageOpts{}, err
	}
	return clients.HcloudServerCreateFromImageOpts{
		Name:           nodeName(cl, name),
		ServerType:     serverType,
		UserData:       string(userData),
		BaseLabels:     map[string]string{clusterLabel: cl.Config.ClusterName},
		FinalizeLabels: map[string]string{roleLabel: "controlplane"},
		ImageTarXzUrl:  "https://github.com/talos-systems/talos/releases/download/v0.14.2/hcloud-amd64.raw.xz",
	}, nil
}

func workerNodeTemplate(cl *cluster.Cluster, serverType string, pool string, name string) (clients.HcloudServerCreateFromImageOpts, error) {
	userData, err := ioutil.ReadFile(path.Join(cl.Dir, "worker.yaml"))
	if err != nil {
		return clients.HcloudServerCreateFromImageOpts{}, err
	}
	return clients.HcloudServerCreateFromImageOpts{
		Name:           nodeName(cl, name),
		ServerType:     serverType,
		UserData:       string(userData),
		BaseLabels:     map[string]string{clusterLabel: cl.Config.ClusterName},
		FinalizeLabels: map[string]string{roleLabel: "worker", poolLabel: pool},
		ImageTarXzUrl:  "https://github.com/talos-systems/talos/releases/download/v0.14.2/hcloud-amd64.raw.xz",
	}, nil
}
