package cmd

import (
	"flag"
	"fmt"
	"os"
	"runtime/debug"

	"github.com/airfocusio/hcloud-talos/internal"
	"github.com/airfocusio/hcloud-talos/internal/utils"
)

type VersionCommandId struct {
	Version FullVersion
}

func (cmdId *VersionCommandId) Name() string {
	return "version"
}

func (cmdId *VersionCommandId) Create() internal.Command {
	return &VersionCommand{Version: cmdId.Version}
}

type VersionCommand struct {
	Version FullVersion
}

func (cmd *VersionCommand) RegisterOpts(flags *flag.FlagSet) {}

func (cmd *VersionCommand) ValidateOpts() error {
	return nil
}

func (cmd *VersionCommand) Run(logger *utils.Logger, dir string) error {
	os.Stdout.Write([]byte(cmd.Version.Version + "\n"))
	return nil
}

type FullVersion struct {
	Version string
	Commit  string
	Date    string
	BuiltBy string
}

func (v FullVersion) ToString() string {
	result := v.Version
	if v.Commit != "" {
		result = fmt.Sprintf("%s\ncommit: %s", result, v.Commit)
	}
	if v.Date != "" {
		result = fmt.Sprintf("%s\nbuilt at: %s", result, v.Date)
	}
	if v.BuiltBy != "" {
		result = fmt.Sprintf("%s\nbuilt by: %s", result, v.BuiltBy)
	}
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Sum != "" {
		result = fmt.Sprintf("%s\nmodule version: %s, checksum: %s", result, info.Main.Version, info.Main.Sum)
	}
	return result
}
