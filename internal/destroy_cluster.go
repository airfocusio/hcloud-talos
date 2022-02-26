package internal

import (
	"fmt"

	"github.com/airfocusio/hcloud-talos/internal/cluster"
	"github.com/airfocusio/hcloud-talos/internal/utils"
	"github.com/hetznercloud/hcloud-go/hcloud"
)

type DestroyClusterOpts struct {
	Force bool
}

func DestroyCluster(logger *utils.Logger, dir string, opts DestroyClusterOpts) error {
	cl := &cluster.Cluster{Dir: dir}
	err := cl.Load(logger)
	if err != nil {
		return err
	}
	logger.Info.Printf("Destroying cluster %q\n", cl.Config.ClusterName)

	if !opts.Force {
		return fmt.Errorf("destroying the cluster must be forced")
	}

	servers, _, err := cl.Client.Server.List(*cl.Ctx, hcloud.ServerListOpts{
		ListOpts: hcloud.ListOpts{
			LabelSelector: clusterLabel + "=" + cl.Config.ClusterName,
		},
	})
	if err != nil {
		logger.Warn.Printf("Error: %v\n", err)
	}
	for _, server := range servers {
		cl.Logger.Info.Printf("Deleting server %d\n", server.ID)
		err := utils.Retry(cl.Logger, func() error {
			_, err := cl.Client.Server.Delete(*cl.Ctx, server)
			return err
		})
		if err != nil {
			logger.Warn.Printf("Error: %v\n", err)
		}
	}

	firewalls, _, err := cl.Client.Firewall.List(*cl.Ctx, hcloud.FirewallListOpts{
		ListOpts: hcloud.ListOpts{
			LabelSelector: clusterLabel + "=" + cl.Config.ClusterName,
		},
	})
	if err != nil {
		logger.Warn.Printf("Error: %v\n", err)
	}
	for _, firewall := range firewalls {
		cl.Logger.Info.Printf("Deleting firewall %d\n", firewall.ID)
		err := utils.Retry(cl.Logger, func() error {
			_, err := cl.Client.Firewall.Delete(*cl.Ctx, firewall)
			return err
		})
		if err != nil {
			logger.Warn.Printf("Error: %v\n", err)
		}
	}

	loadBalancers, _, err := cl.Client.LoadBalancer.List(*cl.Ctx, hcloud.LoadBalancerListOpts{
		ListOpts: hcloud.ListOpts{
			LabelSelector: clusterLabel + "=" + cl.Config.ClusterName,
		},
	})
	if err != nil {
		logger.Warn.Printf("Error: %v\n", err)
	}
	for _, loadBalancer := range loadBalancers {
		cl.Logger.Info.Printf("Deleting load balancer %d\n", loadBalancer.ID)
		err := utils.Retry(cl.Logger, func() error {
			_, err := cl.Client.LoadBalancer.Delete(*cl.Ctx, loadBalancer)
			return err
		})
		if err != nil {
			logger.Warn.Printf("Error: %v\n", err)
		}
	}

	placementGroups, _, err := cl.Client.PlacementGroup.List(*cl.Ctx, hcloud.PlacementGroupListOpts{
		ListOpts: hcloud.ListOpts{
			LabelSelector: clusterLabel + "=" + cl.Config.ClusterName,
		},
	})
	if err != nil {
		logger.Warn.Printf("Error: %v\n", err)
	}
	for _, placementGroup := range placementGroups {
		cl.Logger.Info.Printf("Deleting placement group %d\n", placementGroup.ID)
		err := utils.Retry(cl.Logger, func() error {
			_, err := cl.Client.PlacementGroup.Delete(*cl.Ctx, placementGroup)
			return err
		})
		if err != nil {
			logger.Warn.Printf("Error: %v\n", err)
		}
	}

	networks, _, err := cl.Client.Network.List(*cl.Ctx, hcloud.NetworkListOpts{
		ListOpts: hcloud.ListOpts{
			LabelSelector: clusterLabel + "=" + cl.Config.ClusterName,
		},
	})
	if err != nil {
		logger.Warn.Printf("Error: %v\n", err)
	}
	for _, network := range networks {
		cl.Logger.Info.Printf("Deleting network %d\n", network.ID)
		err := utils.Retry(cl.Logger, func() error {
			_, err := cl.Client.Network.Delete(*cl.Ctx, network)
			return err
		})
		if err != nil {
			logger.Warn.Printf("Error: %v\n", err)
		}
	}

	return nil
}
