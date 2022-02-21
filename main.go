package main

import (
	"os"

	"github.com/airfocusio/hcloud-talos/cmd"
	"github.com/airfocusio/hcloud-talos/internal/utils"
)

// nolint: gochecknoglobals
var (
	version = "dev"
	commit  = ""
	date    = ""
	builtBy = ""
)

func main() {
	logger := utils.NewLogger()
	if err := cmd.Execute(&logger, cmd.FullVersion{Version: version, Commit: commit, Date: date, BuiltBy: builtBy}); err != nil {
		os.Exit(1)
	}
}
