package cmd

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/spf13/cobra"
)

var Version FullVersion

var versionCmd = &cobra.Command{
	Use:  "version",
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		os.Stdout.Write([]byte(Version.ToString() + "\n"))
	},
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
