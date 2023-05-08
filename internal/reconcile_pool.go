package internal

import (
	"fmt"

	"github.com/airfocusio/hcloud-talos/internal/cluster"
	"github.com/airfocusio/hcloud-talos/internal/utils"
	"github.com/hetznercloud/hcloud-go/hcloud"
)

type ReconcilePoolOpts struct {
	ConfigFile     string
	PoolName       string
	NodeNamePrefix string
	NodeCount      int
	ServerType     string
	TalosVersion   string
}

func ReconcilePool(logger *utils.Logger, dir string, opts ReconcilePoolOpts) error {
	cl := &cluster.Cluster{Dir: dir}
	err := cl.Load(opts.ConfigFile, logger)
	if err != nil {
		return err
	}
	logger.Info.Printf("Reconciling pool %s/%s\n", cl.Config.ClusterName, opts.PoolName)
	if opts.PoolName == "" {
		return fmt.Errorf("pool name must not be empty")
	}
	if opts.NodeNamePrefix == "" {
		return fmt.Errorf("node name prefix must not be empty")
	}
	if opts.NodeCount < 0 {
		return fmt.Errorf("node count must not be negative")
	}
	if opts.ServerType == "" {
		return fmt.Errorf("node server type must not be empty")
	}

	for {
		poolServers, _, err := cl.Client.Server.List(*cl.Ctx, hcloud.ServerListOpts{
			ListOpts: hcloud.ListOpts{
				LabelSelector: clusterLabel + "=" + cl.Config.ClusterName + "," + roleLabel + "=worker," + poolLabel + "=" + opts.PoolName,
			},
		})
		if err != nil {
			return err
		}

		nodeCountDiff := len(poolServers) - opts.NodeCount
		if nodeCountDiff < 0 {
			nodeName := opts.NodeNamePrefix + "-%id%"
			_, err = AddNode(logger, cl.Dir, AddNodeOpts{
				ConfigFile:   opts.ConfigFile,
				ServerType:   opts.ServerType,
				NodeName:     nodeName,
				PoolName:     opts.PoolName,
				TalosVersion: opts.TalosVersion,
			})
			if err != nil {
				return err
			}
		} else {
			logger.Debug.Printf("Pool %s/%s already has %d of %d nodes\n", cl.Config.ClusterName, opts.PoolName, len(poolServers), opts.NodeCount)
			break
		}
	}

	return nil
}
