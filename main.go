package main

import (
	"log/slog"

	"github.com/philippheuer/jvm-repo-rebuild-index/pkg/cmd"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
	status  = "clean"
)

// Init Hook
func init() {
	cmd.Version = version
	cmd.CommitHash = commit
	cmd.BuildAt = date
	cmd.RepositoryStatus = status
}

// CLI Main Entrypoint
func main() {
	cmdErr := cmd.Execute()
	if cmdErr != nil {
		slog.Error("cli error", "error", cmdErr)
	}
}
