package main

import (
	"os"

	"github.com/airfocusio/hcloud-talos/cmd"
)

// nolint: gochecknoglobals
var (
	version = "dev"
	commit  = ""
	date    = ""
	builtBy = ""
)

func main() {
	cmd.Version = cmd.FullVersion{
		Version: version,
		Commit:  commit,
		Date:    date,
		BuiltBy: builtBy,
	}
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
