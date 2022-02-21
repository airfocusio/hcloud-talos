package cmd

import (
	"flag"
	"fmt"
	"os"
	"runtime/debug"

	"github.com/airfocusio/hcloud-talos/internal/cmds"
	"github.com/airfocusio/hcloud-talos/internal/utils"
)

func Execute(logger *utils.Logger, version FullVersion) error {
	dir := ""
	flags := flag.NewFlagSet("root", flag.ContinueOnError)
	flags.StringVar(&dir, "dir", "", "")
	if err := flags.Parse(os.Args[1:]); err != nil {
		return err
	}
	if dir == "" {
		dirOut, err := os.Getwd()
		if err != nil {
			return err
		}
		dir = dirOut
	}

	command := ""
	commandArgs := flags.Args()
	if flags.NArg() >= 1 {
		command = flags.Arg(0)
		commandArgs = flags.Args()[1:]
	}

	switch command {
	case "version":
		os.Stdout.Write([]byte(version.Version))
		return nil
	case "bootstrap-cluster":
		cmd := cmds.BootstrapClusterCommand{}
		return RunCommand(logger, &cmd, commandArgs, dir)
	case "destroy-cluster":
		cmd := cmds.DestroyClusterCommand{}
		return RunCommand(logger, &cmd, commandArgs, dir)
	case "apply-manifests":
		cmd := cmds.ApplyManifestsCommand{}
		return RunCommand(logger, &cmd, commandArgs, dir)
	case "add-node":
		cmd := cmds.AddNodeCommand{}
		return RunCommand(logger, &cmd, commandArgs, dir)
	case "":
		err := fmt.Errorf("unknown command %s", command)
		os.Stderr.Write([]byte(err.Error()))
		return err
	default:
		err := fmt.Errorf("unknown command %s", command)
		os.Stderr.Write([]byte(err.Error()))
		return err
	}
}

func RunCommand(logger *utils.Logger, cmd cmds.Command, args []string, dir string) error {
	flags := flag.NewFlagSet("command", flag.ContinueOnError)
	cmd.RegisterOpts(flags)
	if err := flags.Parse(args); err != nil {
		return err
	}
	if err := cmd.ValidateOpts(); err != nil {
		logger.Error.Printf("%v\n", err)
		return err
	}
	if err := cmd.Run(logger, dir); err != nil {
		logger.Error.Printf("%v\n", err)
		return err
	}
	logger.Info.Printf("Done\n")
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
