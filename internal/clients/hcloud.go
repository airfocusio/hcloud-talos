package clients

import (
	"fmt"
	"net"

	"github.com/airfocusio/hcloud-talos/internal/cluster"
	"github.com/airfocusio/hcloud-talos/internal/utils"
	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/hetznercloud/hcloud-go/hcloud/schema"
)

func HcloudEnsureNetwork(cl *cluster.Cluster, tmpl hcloud.NetworkCreateOpts, create bool) (*hcloud.Network, error) {
	network, _, err := cl.Client.Network.Get(*cl.Ctx, tmpl.Name)
	if err != nil {
		return nil, err
	}
	if network != nil {
		return network, nil
	}
	if !create {
		return nil, fmt.Errorf("network %q could not be found", tmpl.Name)
	}

	cl.Logger.Info.Printf("Creating new network %q\n", tmpl.Name)
	network, _, err = cl.Client.Network.Create(*cl.Ctx, tmpl)
	if err != nil {
		return nil, err
	}

	return network, nil
}

func HcloudEnsurePlacementGroup(cl *cluster.Cluster, tmpl hcloud.PlacementGroupCreateOpts, create bool) (*hcloud.PlacementGroup, error) {
	placementGroup, _, err := cl.Client.PlacementGroup.Get(*cl.Ctx, tmpl.Name)
	if err != nil {
		return nil, err
	}
	if placementGroup != nil {
		return placementGroup, nil
	}
	if !create {
		return nil, fmt.Errorf("placement group %q could not be found", tmpl.Name)
	}

	cl.Logger.Info.Printf("Creating new placement group %q\n", tmpl.Name)
	placementGroupResult, _, err := cl.Client.PlacementGroup.Create(*cl.Ctx, tmpl)
	if err != nil {
		return nil, err
	}
	placementGroup = placementGroupResult.PlacementGroup

	return placementGroup, nil
}

func HcloudEnsureLoadBalancer(cl *cluster.Cluster, network *hcloud.Network, tmpl hcloud.LoadBalancerCreateOpts, create bool) (*hcloud.LoadBalancer, error) {
	loadBalancer, _, err := cl.Client.LoadBalancer.Get(*cl.Ctx, tmpl.Name)
	if err != nil {
		return nil, err
	}
	if loadBalancer != nil {
		return loadBalancer, nil
	}
	if !create {
		return nil, fmt.Errorf("load balancer %q could not be found", tmpl.Name)
	}

	targets := tmpl.Targets
	tmpl.Targets = []hcloud.LoadBalancerCreateOptsTarget{}

	cl.Logger.Info.Printf("Creating new load balancer %q\n", tmpl.Name)
	loadBalancerResult, _, err := cl.Client.LoadBalancer.Create(*cl.Ctx, tmpl)
	if err != nil {
		return nil, err
	}
	loadBalancer = loadBalancerResult.LoadBalancer

	err = utils.Retry(cl.Logger, func() error {
		loadBalancer, _, err = cl.Client.LoadBalancer.GetByID(*cl.Ctx, loadBalancer.ID)
		if err != nil {
			return err
		}
		if loadBalancer.PublicNet.IPv4.IP.IsUnspecified() {
			return fmt.Errorf("load balancer does not yet have a public IP")
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	err = utils.Retry(cl.Logger, func() error {
		loadBalancer, _, err = cl.Client.LoadBalancer.GetByID(*cl.Ctx, loadBalancer.ID)
		if err != nil {
			return err
		}
		if tmpl.Network != nil && (len(loadBalancer.PrivateNet) == 0 || loadBalancer.PrivateNet[0].IP.IsUnspecified()) {
			return fmt.Errorf("load balancer does not yet have a private IP")
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	for _, target := range targets {
		err := utils.Retry(cl.Logger, func() error {
			switch target.Type {
			case hcloud.LoadBalancerTargetTypeLabelSelector:
				_, _, err := cl.Client.LoadBalancer.AddLabelSelectorTarget(*cl.Ctx, loadBalancer, hcloud.LoadBalancerAddLabelSelectorTargetOpts{
					Selector:     target.LabelSelector.Selector,
					UsePrivateIP: target.UsePrivateIP,
				})
				return err
			case hcloud.LoadBalancerTargetTypeServer:
				_, _, err := cl.Client.LoadBalancer.AddServerTarget(*cl.Ctx, loadBalancer, hcloud.LoadBalancerAddServerTargetOpts{

					Server:       target.Server.Server,
					UsePrivateIP: target.UsePrivateIP,
				})
				return err
			case hcloud.LoadBalancerTargetTypeIP:
				_, _, err := cl.Client.LoadBalancer.AddIPTarget(*cl.Ctx, loadBalancer, hcloud.LoadBalancerAddIPTargetOpts{
					IP: net.IP(target.IP.IP),
				})
				return err
			default:
				return fmt.Errorf("unknown load balancer target type %s", target.Type)
			}
		})
		if err != nil {
			return nil, err
		}
	}

	return loadBalancer, nil
}

func HcloudEnsureFirewall(cl *cluster.Cluster, tmpl hcloud.FirewallCreateOpts, create bool) (*hcloud.Firewall, error) {
	firewall, _, err := cl.Client.Firewall.Get(*cl.Ctx, tmpl.Name)
	if err != nil {
		return nil, err
	}
	if firewall != nil {
		return firewall, nil
	}
	if !create {
		return nil, fmt.Errorf("firewall %q could not be found", tmpl.Name)
	}

	cl.Logger.Info.Printf("Creating new firewall %q\n", tmpl.Name)
	firewallResult, _, err := cl.Client.Firewall.Create(*cl.Ctx, tmpl)
	if err != nil {
		return nil, err
	}
	firewall = firewallResult.Firewall

	return firewall, nil
}

type HcloudServerCreateFromImageOpts struct {
	Name           string
	ServerType     string
	UserData       string
	BaseLabels     map[string]string
	FinalizeLabels map[string]string
	ImageTarXzUrl  string
}

func HcloudCreateServerFromImage(cl *cluster.Cluster, network *hcloud.Network, placementGroup *hcloud.PlacementGroup, tmpl HcloudServerCreateFromImageOpts) (*hcloud.Server, error) {
	cl.Logger.Info.Printf("Creating new server %q\n", tmpl.Name)
	cl.Logger.Debug.Printf("Generating temporary SSH key\n")
	sshKeyPrivate := SSHKeyPrivate{}
	if err := sshKeyPrivate.Generate(); err != nil {
		return nil, err
	}
	sshKeyPublic, err := sshKeyPrivate.StorePublic()
	if err != nil {
		return nil, err
	}
	cl.Logger.Debug.Printf("Registering temporary SSH key\n")
	sshKey, _, err := cl.Client.SSHKey.Create(*cl.Ctx, hcloud.SSHKeyCreateOpts{
		Name:      tmpl.Name + "-init-" + utils.RandString(8),
		PublicKey: sshKeyPublic,
		Labels:    tmpl.BaseLabels,
	})
	if err != nil {
		return nil, err
	}
	defer func() {
		cl.Logger.Debug.Printf("Removing temporary SSH key\n")
		cl.Client.SSHKey.Delete(*cl.Ctx, sshKey)
	}()

	initImage := "debian-11"
	startAfterCreate := false
	serverRespone, _, err := cl.Client.Server.Create(*cl.Ctx, hcloud.ServerCreateOpts{
		Name: tmpl.Name,
		ServerType: hcloud.ServerTypeFromSchema(schema.ServerType{
			Name: tmpl.ServerType,
		}),
		Image: hcloud.ImageFromSchema(schema.Image{
			Name: &initImage,
		}),
		PlacementGroup: placementGroup,
		Location: &hcloud.Location{
			Name: cl.Config.Hcloud.Location,
		},
		Networks: []*hcloud.Network{
			network,
		},
		StartAfterCreate: &startAfterCreate,
		UserData:         tmpl.UserData,
		SSHKeys:          []*hcloud.SSHKey{sshKey}, // not needed, but without Hetzner sends out server creation email
		Labels:           tmpl.BaseLabels,
	})
	if err != nil {
		return nil, err
	}

	cl.Logger.Debug.Printf("Waiting for server to have private IP\n")
	server := serverRespone.Server
	err = utils.Retry(cl.Logger, func() error {
		server, _, err = cl.Client.Server.GetByID(*cl.Ctx, server.ID)
		if err != nil {
			return err
		}
		if len(server.PrivateNet) == 0 || server.PrivateNet[0].IP.IsUnspecified() {
			return fmt.Errorf("server does not yet have a private IP")
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	cl.Logger.Debug.Printf("Starting server in rescue mode\n")
	err = utils.Retry(cl.Logger, func() error {
		_, _, err = cl.Client.Server.EnableRescue(*cl.Ctx, server, hcloud.ServerEnableRescueOpts{
			Type:    hcloud.ServerRescueTypeLinux64,
			SSHKeys: []*hcloud.SSHKey{sshKey},
		})
		return err
	})
	if err != nil {
		return nil, err
	}
	err = utils.Retry(cl.Logger, func() error {
		_, _, err = cl.Client.Server.Poweron(*cl.Ctx, server)
		return err
	})
	if err != nil {
		return nil, err
	}
	err = utils.RetrySlow(cl.Logger, func() error {
		_, err := sshKeyPrivate.Execute(server.PublicNet.IPv4.IP.String(), 22, "true")
		return err
	})
	if err != nil {
		return nil, err
	}

	cl.Logger.Debug.Printf("Applying image\n")
	err = utils.Retry(cl.Logger, func() error {
		_, err := sshKeyPrivate.Execute(server.PublicNet.IPv4.IP.String(), 22, fmt.Sprintf(`
			cd /tmp
			wget -O /tmp/image.xz %s
			xz -d -c /tmp/image.xz | dd of=/dev/sda && sync
		`, tmpl.ImageTarXzUrl))
		return err
	})
	if err != nil {
		return nil, err
	}

	cl.Logger.Debug.Printf("Shutting down server\n")
	err = utils.Retry(cl.Logger, func() error {
		_, _, err = cl.Client.Server.Shutdown(*cl.Ctx, server)
		return err
	})
	if err != nil {
		return nil, err
	}
	err = utils.Retry(cl.Logger, func() error {
		server, _, err := cl.Client.Server.GetByID(*cl.Ctx, server.ID)
		if err != nil {
			return err
		}
		if server.Status != hcloud.ServerStatusOff {
			return fmt.Errorf("server is not yet turned off")
		}
		return nil
	})

	cl.Logger.Debug.Printf("Configuring server labels\n")
	baseAndFinalizeLabels := map[string]string{}
	for k, v := range tmpl.BaseLabels {
		baseAndFinalizeLabels[k] = v
	}
	for k, v := range tmpl.FinalizeLabels {
		baseAndFinalizeLabels[k] = v
	}
	err = utils.Retry(cl.Logger, func() error {
		server, _, err = cl.Client.Server.Update(*cl.Ctx, server, hcloud.ServerUpdateOpts{
			Labels: baseAndFinalizeLabels,
		})
		return err
	})
	if err != nil {
		return nil, err
	}

	cl.Logger.Debug.Printf("Starting server\n")
	err = utils.Retry(cl.Logger, func() error {
		_, _, err = cl.Client.Server.Poweron(*cl.Ctx, server)
		return err
	})
	if err != nil {
		return nil, err
	}

	return server, nil
}
