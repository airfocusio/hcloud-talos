package cmds

import (
	"flag"

	"github.com/airfocusio/hcloud-talos/internal/utils"
)

type Command interface {
	RegisterOpts(flags *flag.FlagSet)
	ValidateOpts() error
	Run(logger *utils.Logger, dir string) error
}
