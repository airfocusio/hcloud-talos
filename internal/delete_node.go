package internal

import (
	"fmt"
	"net"

	"github.com/airfocusio/hcloud-talos/internal/clients"
	"github.com/airfocusio/hcloud-talos/internal/cluster"
	"github.com/airfocusio/hcloud-talos/internal/utils"
	"github.com/hetznercloud/hcloud-go/hcloud"
)

type DeleteNodeOpts struct {
	NodeName   string
	KeepServer bool
	Force      bool
}

func DeleteNode(logger *utils.Logger, dir string, opts DeleteNodeOpts) error {
	cl := &cluster.Cluster{Dir: dir}
	err := cl.Load(logger)
	if err != nil {
		return err
	}
	logger.Info.Printf("Deleting node %s/%s\n", cl.Config.ClusterName, opts.NodeName)
	if opts.NodeName == "" {
		return fmt.Errorf("node name must not be empty")
	}
	if !opts.Force {
		return fmt.Errorf("deleting a node must be forced")
	}

	serverName := nodeName(cl, opts.NodeName)
	server, _, err := cl.Client.Server.Get(*cl.Ctx, serverName)
	if err != nil {
		return err
	}
	if server == nil {
		return fmt.Errorf("server %q could not be found", serverName)
	}
	if len(server.PrivateNet) == 0 || server.PrivateNet[0].IP.Equal(net.IP{}) {
		return fmt.Errorf("server %q private IP could not be determined", serverName)
	}
	serverIP := server.PrivateNet[0].IP

	logger.Debug.Printf("Resetting talos\n")
	err = utils.Retry(cl.Logger, func() error {
		_, err := TalosReset(cl, serverIP)
		return err
	})
	if err != nil {
		return err
	}

	logger.Debug.Printf("Waiting for server to shut down talos\n")
	err = utils.RetrySlow(cl.Logger, func() error {
		server, _, err := cl.Client.Server.GetByID(*cl.Ctx, server.ID)
		if err != nil {
			return err
		}
		if server != nil && server.Status != hcloud.ServerStatusOff {
			return fmt.Errorf("server is not yet shut down")
		}
		return nil
	})
	if err != nil {
		logger.Warn.Printf("Server could not be shut down\n")
	}

	err = utils.Retry(cl.Logger, func() error {
		err := clients.KubernetesDeleteNode(cl, serverName)
		return err
	})
	if err != nil {
		return err
	}

	if !opts.KeepServer {
		err = utils.Retry(cl.Logger, func() error {
			_, err := cl.Client.Server.Delete(*cl.Ctx, server)
			return err
		})
		if err != nil {
			return err
		}
	}

	return nil
}
