package cmds

import (
	"flag"
	"fmt"

	"github.com/airfocusio/hcloud-talos/internal"
	"github.com/airfocusio/hcloud-talos/internal/utils"
	"github.com/hetznercloud/hcloud-go/hcloud"
)

type DestroyClusterCommandId struct{}

func (cmdId *DestroyClusterCommandId) Name() string {
	return "destroy-cluster"
}

func (cmdId *DestroyClusterCommandId) Create() Command {
	return &DestroyClusterCommand{}
}

type DestroyClusterCommand struct {
	Force bool
}

func (cmd *DestroyClusterCommand) RegisterOpts(flags *flag.FlagSet) {
	flags.BoolVar(&cmd.Force, "force", false, "")
}

func (cmd *DestroyClusterCommand) ValidateOpts() error {
	if !cmd.Force {
		return fmt.Errorf("must be forced")
	}
	return nil
}

func (cmd *DestroyClusterCommand) Run(logger *utils.Logger, dir string) error {
	ctx := &internal.Context{Dir: dir}
	err := ctx.Load(logger)
	if err != nil {
		return err
	}

	servers, _, err := ctx.Client.Server.List(*ctx.Ctx, hcloud.ServerListOpts{
		ListOpts: hcloud.ListOpts{
			LabelSelector: internal.HcloudMarkerLabel + "=" + ctx.Config.ClusterName,
		},
	})
	if err != nil {
		logger.Warn.Printf("Error: %v\n", err)
	}
	for _, server := range servers {
		ctx.Logger.Info.Printf("Deleting server %d\n", server.ID)
		err := utils.Retry(ctx.Logger, func() error {
			_, err := ctx.Client.Server.Delete(*ctx.Ctx, server)
			return err
		})
		if err != nil {
			logger.Warn.Printf("Error: %v\n", err)
		}
	}

	firewalls, _, err := ctx.Client.Firewall.List(*ctx.Ctx, hcloud.FirewallListOpts{
		ListOpts: hcloud.ListOpts{
			LabelSelector: internal.HcloudMarkerLabel + "=" + ctx.Config.ClusterName,
		},
	})
	if err != nil {
		logger.Warn.Printf("Error: %v\n", err)
	}
	for _, firewall := range firewalls {
		ctx.Logger.Info.Printf("Deleting firewall %d\n", firewall.ID)
		err := utils.Retry(ctx.Logger, func() error {
			_, err := ctx.Client.Firewall.Delete(*ctx.Ctx, firewall)
			return err
		})
		if err != nil {
			logger.Warn.Printf("Error: %v\n", err)
		}
	}

	loadBalancers, _, err := ctx.Client.LoadBalancer.List(*ctx.Ctx, hcloud.LoadBalancerListOpts{
		ListOpts: hcloud.ListOpts{
			LabelSelector: internal.HcloudMarkerLabel + "=" + ctx.Config.ClusterName,
		},
	})
	if err != nil {
		logger.Warn.Printf("Error: %v\n", err)
	}
	for _, loadBalancer := range loadBalancers {
		ctx.Logger.Info.Printf("Deleting load balancer %d\n", loadBalancer.ID)
		err := utils.Retry(ctx.Logger, func() error {
			_, err := ctx.Client.LoadBalancer.Delete(*ctx.Ctx, loadBalancer)
			return err
		})
		if err != nil {
			logger.Warn.Printf("Error: %v\n", err)
		}
	}

	placementGroups, _, err := ctx.Client.PlacementGroup.List(*ctx.Ctx, hcloud.PlacementGroupListOpts{
		ListOpts: hcloud.ListOpts{
			LabelSelector: internal.HcloudMarkerLabel + "=" + ctx.Config.ClusterName,
		},
	})
	if err != nil {
		logger.Warn.Printf("Error: %v\n", err)
	}
	for _, placementGroup := range placementGroups {
		ctx.Logger.Info.Printf("Deleting placement group %d\n", placementGroup.ID)
		err := utils.Retry(ctx.Logger, func() error {
			_, err := ctx.Client.PlacementGroup.Delete(*ctx.Ctx, placementGroup)
			return err
		})
		if err != nil {
			logger.Warn.Printf("Error: %v\n", err)
		}
	}

	networks, _, err := ctx.Client.Network.List(*ctx.Ctx, hcloud.NetworkListOpts{
		ListOpts: hcloud.ListOpts{
			LabelSelector: internal.HcloudMarkerLabel + "=" + ctx.Config.ClusterName,
		},
	})
	if err != nil {
		logger.Warn.Printf("Error: %v\n", err)
	}
	for _, network := range networks {
		ctx.Logger.Info.Printf("Deleting network %d\n", network.ID)
		err := utils.Retry(ctx.Logger, func() error {
			_, err := ctx.Client.Network.Delete(*ctx.Ctx, network)
			return err
		})
		if err != nil {
			logger.Warn.Printf("Error: %v\n", err)
		}
	}

	return nil
}
