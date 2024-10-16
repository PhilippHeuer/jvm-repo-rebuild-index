package cmd

import (
	"log/slog"
	"os"

	"github.com/philippheuer/jvm-repo-rebuild-index/pkg/httpapi"
	"github.com/spf13/cobra"
)

func serveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "print version information",
		Run: func(cmd *cobra.Command, args []string) {
			// flags
			port, _ := cmd.Flags().GetInt("port")
			indexDir, _ := cmd.Flags().GetString("index-dir")
			indexURL, _ := cmd.Flags().GetString("index-url")
			if indexDir == "" && indexURL == "" {
				slog.Error("Either index-dir or index-url must be set")
				return
			}

			// start server
			err := httpapi.Serve(port, indexDir, indexURL)
			if err != nil {
				slog.Error("Error starting server", "err", err)
				os.Exit(1)
			}
		},
	}

	cmd.Flags().IntP("port", "p", 8080, "Port")
	cmd.Flags().String("index-dir", "", "Index directory (for local index)")
	cmd.Flags().String("index-url", "https://philippheuer.github.io/jvm-repo-rebuild-index", "Index URL (as proxy for remote index)")

	return cmd
}
