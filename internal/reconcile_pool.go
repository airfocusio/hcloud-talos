package internal

import (
	"fmt"
	"strings"

	"github.com/airfocusio/hcloud-talos/internal/cluster"
	"github.com/airfocusio/hcloud-talos/internal/utils"
	"github.com/hetznercloud/hcloud-go/hcloud"
)

type ReconcilePoolOpts struct {
	PoolName       string
	NodeNamePrefix string
	NodeCount      int
	ServerType     string
	Force          bool
}

func ReconcilePool(logger *utils.Logger, dir string, opts ReconcilePoolOpts) error {
	cl := &cluster.Cluster{Dir: dir}
	err := cl.Load(logger)
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
			nodeName, err := findNextNodeName(cl, opts.NodeNamePrefix, poolServers)
			if err != nil {
				return err
			}
			_, err = AddNode(logger, cl.Dir, AddNodeOpts{
				ServerType: opts.ServerType,
				NodeName:   nodeName,
				PoolName:   opts.PoolName,
			})
			if err != nil {
				return err
			}
		} else if nodeCountDiff > 0 {
			server := poolServers[len(poolServers)-1]
			err := DeleteNode(logger, cl.Dir, DeleteNodeOpts{
				NodeName: strings.TrimPrefix(server.Name, cl.Config.ClusterName+"-"),
				Force:    opts.Force,
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

func findNextNodeName(cl *cluster.Cluster, nodeNamePrefix string, poolServers []*hcloud.Server) (string, error) {
	for i := 1; i <= 1000; i++ {
		exists := false
		proposedNodeName := fmt.Sprintf("%s-%d", nodeNamePrefix, i)
		for _, server := range poolServers {
			if server.Name == nodeName(cl, proposedNodeName) {
				exists = true
				break
			}
		}
		if !exists {
			return proposedNodeName, nil
		}
	}
	return "", fmt.Errorf("unable to find next node name")
}
