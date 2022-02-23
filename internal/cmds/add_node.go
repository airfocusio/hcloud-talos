package cmds

import (
	"flag"
	"fmt"

	"github.com/airfocusio/hcloud-talos/internal"
	"github.com/airfocusio/hcloud-talos/internal/utils"
	"github.com/hetznercloud/hcloud-go/hcloud"
)

type AddNodeCommandId struct{}

func (cmdId *AddNodeCommandId) Name() string {
	return "add-node"
}

func (cmdId *AddNodeCommandId) Create() Command {
	return &AddNodeCommand{}
}

type AddNodeCommand struct {
	NodeName       string
	NodeServerType string
	Controlplane   bool
}

var _ Command = (*AddNodeCommand)(nil)

func (cmd *AddNodeCommand) RegisterOpts(flags *flag.FlagSet) {
	flags.StringVar(&cmd.NodeName, "node-name", "", "")
	flags.StringVar(&cmd.NodeServerType, "node-server-type", "cx21", "")
	flags.BoolVar(&cmd.Controlplane, "controlplane", false, "")
}

func (cmd *AddNodeCommand) ValidateOpts() error {
	if cmd.NodeName == "" {
		return fmt.Errorf("node name must not be empty")
	}
	if cmd.NodeServerType == "" {
		return fmt.Errorf("node server type must not be empty")
	}
	return nil
}

func (cmd *AddNodeCommand) Run(logger *utils.Logger, dir string) error {
	ctx := &internal.Context{Dir: dir}
	err := ctx.Load(logger)
	if err != nil {
		ctx.Logger.Error.Printf("Error: %v\n", err)
		return err
	}

	network, err := internal.HcloudEnsureNetwork(ctx)
	if err != nil {
		ctx.Logger.Error.Printf("Error: %v\n", err)
		return err
	}

	placementGroup, err := internal.HcloudEnsurePlacementGroup(ctx)
	if err != nil {
		ctx.Logger.Error.Printf("Error: %v\n", err)
		return err
	}

	var server *hcloud.Server
	if cmd.Controlplane {
		server, err = internal.TalosCreateControlplane(ctx, network, placementGroup, cmd.NodeName, cmd.NodeServerType)
	} else {
		server, err = internal.TalosCreateWorker(ctx, network, placementGroup, cmd.NodeName, cmd.NodeServerType)
	}
	if err != nil {
		ctx.Logger.Error.Printf("Error: %v\n", err)
		return err
	}

	err = utils.RetrySlow(ctx.Logger, func() error {
		nodes, err := internal.KubernetesListNodes(ctx)
		if err != nil {
			return err
		}
		for _, node := range nodes.Items {
			if node.Name == server.Name {
				return nil
			}
		}
		return fmt.Errorf("node not yet available")
	})
	if err != nil {
		ctx.Logger.Error.Printf("Error: %v\n", err)
		return err
	}

	ctx.Logger.Info.Printf("Done\n")
	return nil
}
