package internal

import (
	"flag"

	"github.com/airfocusio/hcloud-talos/internal/utils"
)

type CommandId interface {
	Name() string
	Create() Command
}

type Command interface {
	RegisterOpts(flags *flag.FlagSet)
	ValidateOpts() error
	Run(logger *utils.Logger, dir string) error
}
