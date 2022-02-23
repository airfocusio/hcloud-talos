package internal

import (
	"flag"
	"fmt"
	"strings"

	"github.com/airfocusio/hcloud-talos/internal/cluster"
	"github.com/airfocusio/hcloud-talos/internal/utils"
	"github.com/hetznercloud/hcloud-go/hcloud"
)

type ReconcilePoolCommandId struct{}

func (cmdId *ReconcilePoolCommandId) Name() string {
	return "reconcile-pool"
}

func (cmdId *ReconcilePoolCommandId) Create() Command {
	return &ReconcilePoolCommand{}
}

type ReconcilePoolCommand struct {
	PoolName       string
	NodeNamePrefix string
	NodeCount      int
	ServerType     string
	Force          bool
}

func (cmd *ReconcilePoolCommand) RegisterOpts(flags *flag.FlagSet) {
	flags.StringVar(&cmd.PoolName, "pool-name", "", "")
	flags.StringVar(&cmd.NodeNamePrefix, "node-name-prefix", "", "")
	flags.StringVar(&cmd.ServerType, "server-type", "cx21", "")
	flags.IntVar(&cmd.NodeCount, "node-count", -1, "")
	flags.BoolVar(&cmd.Force, "force", false, "")
}

func (cmd *ReconcilePoolCommand) ValidateOpts() error {
	if cmd.PoolName == "" {
		return fmt.Errorf("pool name must not be empty")
	}
	if cmd.NodeNamePrefix == "" {
		return fmt.Errorf("node name prefix must not be empty")
	}
	if cmd.NodeCount < 0 {
		return fmt.Errorf("node count must not be negative")
	}
	if cmd.ServerType == "" {
		return fmt.Errorf("node server type must not be empty")
	}
	return nil
}

func (cmd *ReconcilePoolCommand) Run(logger *utils.Logger, dir string) error {
	cl := &cluster.Cluster{Dir: dir}
	err := cl.Load(logger)
	if err != nil {
		return err
	}
	logger.Info.Printf("Reconciling pool %s/%s\n", cl.Config.ClusterName, cmd.PoolName)

	for {
		poolServers, _, err := cl.Client.Server.List(*cl.Ctx, hcloud.ServerListOpts{
			ListOpts: hcloud.ListOpts{
				LabelSelector: clusterLabel + "=" + cl.Config.ClusterName + "," + roleLabel + "=worker," + poolLabel + "=" + cmd.PoolName,
			},
		})
		if err != nil {
			return err
		}

		nodeCountDiff := len(poolServers) - cmd.NodeCount
		if nodeCountDiff < 0 {
			nodeName, err := findNextNodeName(cl, cmd.NodeNamePrefix, poolServers)
			if err != nil {
				return err
			}
			cmd := AddNodeCommand{
				ServerType: cmd.ServerType,
				NodeName:   nodeName,
				PoolName:   cmd.PoolName,
			}
			err = cmd.Run(logger, cl.Dir)
			if err != nil {
				return err
			}
		} else if nodeCountDiff > 0 {
			server := poolServers[len(poolServers)-1]
			cmd := DeleteNodeCommand{
				NodeName: strings.TrimPrefix(server.Name, cl.Config.ClusterName+"-"),
				Force:    cmd.Force,
			}
			err := cmd.Run(logger, cl.Dir)
			if err != nil {
				return err
			}
		} else {
			logger.Debug.Printf("Pool %s/%s already has %d of %d nodes\n", cl.Config.ClusterName, cmd.PoolName, len(poolServers), cmd.NodeCount)
			break
		}
	}

	return nil
}

func findNextNodeName(cl *cluster.Cluster, nodeNamePrefix string, poolServers []*hcloud.Server) (string, error) {
	for i := 1; i <= 1000; i++ {
		exists := false
		proposedNodeName := fmt.Sprintf("%s%d", nodeNamePrefix, i)
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
