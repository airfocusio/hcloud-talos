package cmd

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/airfocusio/hcloud-talos/internal"
	"github.com/airfocusio/hcloud-talos/internal/utils"
)

func Execute(version FullVersion) error {
	commandIds := []internal.CommandId{
		&VersionCommandId{Version: version},
		&internal.BootstrapClusterCommandId{},
		&internal.DestroyClusterCommandId{},
		&internal.ApplyManifestsCommandId{},
		&internal.AddNodeCommandId{},
		&internal.DeleteNodeCommandId{},
	}

	availableCommands := []string{}
	for _, commandId := range commandIds {
		availableCommands = append(availableCommands, commandId.Name())
	}
	sort.Strings(availableCommands)

	dir := ""
	verbose := false
	flags := flag.NewFlagSet("root", flag.ContinueOnError)
	flags.StringVar(&dir, "dir", "", "")
	flags.BoolVar(&verbose, "verbose", false, "")
	if err := flags.Parse(os.Args[1:]); err != nil {
		return err
	}
	logger := utils.NewLogger(verbose)
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

	if command == "" {
		err := fmt.Errorf("missing command: available commands are %s", strings.Join(availableCommands, ", "))
		os.Stderr.Write([]byte(err.Error() + "\n"))
		return err
	}

	for _, commandId := range commandIds {
		if commandId.Name() == command {
			cmd := commandId.Create()
			return RunCommand(&logger, cmd, commandArgs, dir)
		}
	}

	err := fmt.Errorf("unknown command %s: available commands are %s", command, strings.Join(availableCommands, ", "))
	os.Stderr.Write([]byte(err.Error() + "\n"))
	return err
}

func RunCommand(logger *utils.Logger, cmd internal.Command, args []string, dir string) error {
	flags := flag.NewFlagSet("command", flag.ContinueOnError)
	cmd.RegisterOpts(flags)
	if err := flags.Parse(args); err != nil {
		return err
	}
	if err := cmd.ValidateOpts(); err != nil {
		logger.Error.Printf("%v\n", err)
		return err
	}
	logger.Debug.Printf("Starting\n")
	if err := cmd.Run(logger, dir); err != nil {
		logger.Error.Printf("%v\n", err)
		return err
	}
	logger.Debug.Printf("Done\n")
	return nil
}
