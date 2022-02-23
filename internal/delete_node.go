package internal

import (
	"flag"
	"fmt"

	"github.com/airfocusio/hcloud-talos/internal/clients"
	"github.com/airfocusio/hcloud-talos/internal/cluster"
	"github.com/airfocusio/hcloud-talos/internal/utils"
	"github.com/hetznercloud/hcloud-go/hcloud"
)

type DeleteNodeCommandId struct{}

func (cmdId *DeleteNodeCommandId) Name() string {
	return "delete-node"
}

func (cmdId *DeleteNodeCommandId) Create() Command {
	return &DeleteNodeCommand{}
}

type DeleteNodeCommand struct {
	NodeName   string
	KeepServer bool
	Force      bool
}

var _ Command = (*DeleteNodeCommand)(nil)

func (cmd *DeleteNodeCommand) RegisterOpts(flags *flag.FlagSet) {
	flags.StringVar(&cmd.NodeName, "node-name", "", "")
	flags.BoolVar(&cmd.KeepServer, "keep-server", false, "")
	flags.BoolVar(&cmd.Force, "force", false, "")
}

func (cmd *DeleteNodeCommand) ValidateOpts() error {
	if cmd.NodeName == "" {
		return fmt.Errorf("node name must not be empty")
	}
	return nil
}

func (cmd *DeleteNodeCommand) Run(logger *utils.Logger, dir string) error {
	cl := &cluster.Cluster{Dir: dir}
	err := cl.Load(logger)
	if err != nil {
		return err
	}

	if !cmd.Force {
		return fmt.Errorf("deleting a node must be forced")
	}

	serverName := nodeName(cl, cmd.NodeName)
	server, _, err := cl.Client.Server.Get(*cl.Ctx, serverName)
	if err != nil {
		return err
	}
	if server == nil {
		return fmt.Errorf("server %q could not be found", serverName)
	}
	if len(server.PrivateNet) == 0 || server.PrivateNet[0].IP.IsUnspecified() {
		return fmt.Errorf("server %q private IP could not be determined", serverName)
	}
	serverIP := server.PrivateNet[0].IP

	err = utils.Retry(cl.Logger, func() error {
		_, err := clients.TalosReset(cl, serverIP.String())
		return err
	})
	if err != nil {
		return err
	}

	err = utils.RetrySlow(cl.Logger, func() error {
		server, _, err := cl.Client.Server.GetByID(*cl.Ctx, server.ID)
		if err != nil {
			return err
		}
		if server.Status != hcloud.ServerStatusOff {
			return fmt.Errorf("server is not yet turned off")
		}
		return nil
	})
	if err != nil {
		return err
	}
	if err != nil {
		return err
	}

	err = utils.Retry(cl.Logger, func() error {
		err := clients.KubernetesDeleteNode(cl, serverName)
		return err
	})
	if err != nil {
		return err
	}

	if !cmd.KeepServer {
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
