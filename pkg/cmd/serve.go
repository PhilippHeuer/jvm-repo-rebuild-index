package cmd

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/philippheuer/reproducible-central-index/pkg/badge"
	"github.com/philippheuer/reproducible-central-index/pkg/service"
	"github.com/philippheuer/reproducible-central-index/pkg/util"
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

			// config
			e := echo.New()
			e.HideBanner = true
			e.HidePort = true

			// middlewares
			e.Use(middleware.Recover())
			e.Use(middleware.Logger())

			// services
			handlerStruct := handlers{
				lookupService: service.NewDependencyLookupService(indexDir, indexURL),
			}

			// handlers
			e.GET("/health", func(c echo.Context) error {
				return c.JSON(http.StatusOK, "OK")
			})
			e.GET("/v1/badge/reproducible/maven/:coordinate/:version", handlerStruct.dependencyBadgeHandler)
			e.GET("/v1/badge/reproducible/maven/:repository/:coordinate/:version", handlerStruct.dependencyBadgeHandler)
			e.GET("/v1/redirect/reproducible/maven/:coordinate/:version", handlerStruct.redirectHandler)
			e.GET("/v1/redirect/reproducible/maven/:repository/:coordinate/:version", handlerStruct.redirectHandler)

			// start
			startErr := e.Start(fmt.Sprintf(":%d", port))
			if startErr != nil {
				if startErr.Error() == "http: Server closed" {
					return
				}
				slog.Error("Error starting server: %s", startErr)
			}
		},
	}

	cmd.Flags().IntP("port", "p", 8080, "Port")
	cmd.Flags().String("index-dir", "", "Index directory (for local index)")
	cmd.Flags().String("index-url", "https://philippheuer.github.io/reproducible-central-index", "Index URL (as proxy for remote index)")

	return cmd
}

type handlers struct {
	lookupService service.DependencyLookupService
}

func (h handlers) dependencyBadgeHandler(c echo.Context) error {
	repository := c.Param("repository")
	coordinate := c.Param("coordinate")
	artifactVersion := c.Param("version")
	theme := c.QueryParam("theme")
	if repository == "" {
		repository = "repo.maven.apache.org/maven2" // default to Maven Central
	}
	if coordinate == "" {
		return c.JSON(http.StatusBadRequest, "param coordinate is required")
	}
	if artifactVersion == "" {
		return c.JSON(http.StatusBadRequest, "param version is required")
	}
	if !util.IsValidMavenCoordinate(coordinate) {
		return c.JSON(http.StatusBadRequest, "query param coordinate is not a valid maven coordinate")
	}

	// lookup
	data, err := h.lookupService.LookupDependencyMetadata(repository, coordinate)
	if err != nil {
		if errors.Is(err, service.ErrRepositoryNotFound) {
			return c.JSON(http.StatusOK, badge.NewDependencyBadge("repository not configured", badge.Error, theme))
		} else if errors.Is(err, service.ErrDependencyNotFound) {
			return c.JSON(http.StatusOK, badge.NewDependencyBadge("not configured", badge.Error, theme))
		}

		slog.Error("Error looking up dependency metadata: %s", err)
		return c.JSON(http.StatusInternalServerError, "internal server error")
	}

	// search version in data
	version, ok := data.Versions[artifactVersion]
	if !ok {
		return c.JSON(http.StatusOK, badge.NewDependencyBadge("pending verification", badge.Warning, theme))
	}

	return c.JSON(http.StatusOK, badge.NewDependencyBadge(
		fmt.Sprintf("%d/%d ok", version.ProjectReproducibleFiles, version.ProjectReproducibleFiles+version.ProjectNonReproducibleFiles),
		util.Ternary(version.Reproducible, badge.Success, badge.Error),
		theme),
	)
}

func (h handlers) redirectHandler(c echo.Context) error {
	repository := c.Param("repository")
	coordinate := c.Param("coordinate")
	if repository == "" {
		repository = "repo.maven.apache.org/maven2" // default to Maven Central
	}
	if coordinate == "" {
		return c.JSON(http.StatusBadRequest, "query param coordinate is required")
	}
	if !util.IsValidMavenCoordinate(coordinate) {
		return c.JSON(http.StatusBadRequest, "query param coordinate is not a valid maven coordinate")
	}

	// lookup
	data, err := h.lookupService.LookupDependencyMetadata(repository, coordinate)
	if err != nil {
		if errors.Is(err, service.ErrRepositoryNotFound) {
			return c.JSON(http.StatusBadRequest, "repository not configured")
		} else if errors.Is(err, service.ErrDependencyNotFound) {
			return c.JSON(http.StatusBadRequest, "artifact not configured")
		}

		slog.Error("Error looking up dependency metadata: %s", err)
		return c.JSON(http.StatusInternalServerError, "internal server error")
	}

	// redirect
	return c.Redirect(http.StatusFound, data.ReproducibleOverviewURL)
}
