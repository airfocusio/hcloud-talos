package internal

import (
	"fmt"

	"github.com/airfocusio/hcloud-talos/internal/clients"
	"github.com/airfocusio/hcloud-talos/internal/cluster"
	"github.com/airfocusio/hcloud-talos/internal/utils"
)

type BootstrapClusterOpts struct {
	ConfigFile                     string
	ClusterName                    string
	NodeName                       string
	ServerType                     string
	Location                       string
	NetworkZone                    string
	Token                          string
	NoFirewall                     bool
	NoTalosKubespan                bool
	NoHcloudCloudControllerManager bool
	NoHcloudCsiDriver              bool
	TalosVersion                   string
	KubernetesVersion              string
}

func BootstrapCluster(logger *utils.Logger, dir string, opts BootstrapClusterOpts) error {
	cl := &cluster.Cluster{Dir: dir}
	err := cl.Create(logger, opts.ClusterName, opts.Location, opts.NetworkZone, opts.Token)
	if err != nil {
		return err
	}
	err = cl.Save(opts.ConfigFile)
	if err != nil {
		return err
	}
	logger.Info.Printf("Bootstrapping cluster %s (talos %s, kubernetes %s)\n", cl.Config.ClusterName, opts.TalosVersion, opts.KubernetesVersion)
	if opts.ClusterName == "" {
		return fmt.Errorf("cluster name must not be empty")
	}
	if opts.NodeName == "" {
		return fmt.Errorf("node name must not be empty")
	}
	if opts.ServerType == "" {
		return fmt.Errorf("node server type must not be empty")
	}
	if opts.Location == "" {
		return fmt.Errorf("location must not be empty")
	}
	if opts.NetworkZone == "" {
		return fmt.Errorf("network zone must not be empty")
	}
	if opts.Token == "" {
		return fmt.Errorf("token must not be empty")
	}
	if opts.TalosVersion == "" {
		return fmt.Errorf("talos version must not be empty")
	}
	if opts.KubernetesVersion == "" {
		return fmt.Errorf("kubernetes version must not be empty")
	}

	network, err := clients.HcloudEnsureNetwork(cl, nodeNetworkTemplate(cl), true)
	if err != nil {
		return err
	}

	placementGroup, err := clients.HcloudEnsurePlacementGroup(cl, nodePlacementGroupTemplate(cl), true)
	if err != nil {
		return err
	}

	controlplaneLoadBalancer, err := clients.HcloudEnsureLoadBalancer(cl, network, controlplaneLoadBalanacerTemplate(cl, network), true)
	if err != nil {
		return err
	}

	if !opts.NoFirewall {
		_, err = clients.HcloudEnsureFirewall(cl, nodeFirewallTemplate(cl, network), true)
		if err != nil {
			return err
		}
	}

	_, err = TalosGenConfig(cl, network, opts.ClusterName, controlplaneLoadBalancer.PublicNet.IPv4.IP, opts.KubernetesVersion, !opts.NoTalosKubespan)
	if err != nil {
		return err
	}

	controlplaneNodeTemplate, err := controlplaneNodeTemplate(cl, opts.ServerType, opts.NodeName, opts.TalosVersion)
	if err != nil {
		return err
	}
	controlplaneServer, err := clients.HcloudCreateServerFromImage(cl, network, placementGroup, controlplaneNodeTemplate)
	if err != nil {
		return err
	}
	controlplaneServerPrivateIP := controlplaneServer.PrivateNet[0].IP

	err = utils.RetrySlow(logger, func() error {
		_, err := TalosBootstrap(cl, controlplaneServerPrivateIP)
		return err
	})
	if err != nil {
		return err
	}

	err = utils.Retry(logger, func() error {
		_, err := TalosKubeconfig(cl, controlplaneServerPrivateIP)
		return err
	})
	if err != nil {
		return err
	}

	err = clients.KubernetesWaitNodeRegistered(cl, controlplaneServer.Name)
	if err != nil {
		return err
	}

	err = utils.Retry(logger, func() error {
		return TalosPatchFlannelDaemonSet(cl, `
			[
				{
					"op": "add",
					"path": "/spec/template/spec/containers/0/args/-",
					"value": "--iface=eth1"
				}
			]
		`)
	})
	if err != nil {
		return err
	}

	err = ApplyManifests(logger, dir, ApplyManifestsOpts{
		ConfigFile:                     opts.ConfigFile,
		NoHcloudCloudControllerManager: opts.NoHcloudCloudControllerManager,
		NoHcloudCsiDriver:              opts.NoHcloudCsiDriver,
	})
	if err != nil {
		return err
	}

	return nil
}
