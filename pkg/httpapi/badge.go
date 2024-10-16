package httpapi

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/philippheuer/jvm-repo-rebuild-index/pkg/badge"
	"github.com/philippheuer/jvm-repo-rebuild-index/pkg/service"
	"github.com/philippheuer/jvm-repo-rebuild-index/pkg/util"
)

func (h handlers) projectBadgeHandler(c echo.Context) error {
	registry := c.Param("registry")
	coordinate := c.Param("coordinate")
	artifactVersion := c.Param("version")
	theme := c.QueryParam("theme")
	if registry == "" {
		registry = "repo.maven.apache.org/maven2" // default to Maven Central
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
	data, err := h.lookupService.LookupProject(registry, coordinate)
	if err != nil {
		if errors.Is(err, service.ErrRepositoryNotFound) {
			return c.JSON(http.StatusOK, badge.NewDependencyBadge("repository not configured", badge.Error, theme))
		} else if errors.Is(err, service.ErrDependencyNotFound) {
			return c.JSON(http.StatusOK, badge.NewDependencyBadge("not configured", badge.Error, theme))
		}

		slog.Error("Error looking up dependency metadata", "err", err)
		return c.JSON(http.StatusInternalServerError, "internal server error")
	}

	// support "latest" as version
	if artifactVersion == "latest" {
		artifactVersion = data.Latest
	}

	// search version in data
	version, ok := data.Versions[artifactVersion]
	if !ok {
		return c.JSON(http.StatusOK, badge.NewDependencyBadge("pending verification", badge.Warning, theme))
	}

	return c.JSON(http.StatusOK, badge.NewDependencyBadge(
		fmt.Sprintf("%s - %d/%d ok", artifactVersion, version.FileStats.TotalReproducibleFiles, version.FileStats.TotalReproducibleFiles+version.FileStats.TotalNonReproducibleFiles),
		util.Ternary(version.Reproducible, badge.Success, badge.Error),
		theme),
	)
}

func (h handlers) dependencyBadgeHandler(c echo.Context) error {
	registry := c.Param("registry")
	coordinate := c.Param("coordinate")
	artifactVersion := c.Param("version")
	theme := c.QueryParam("theme")
	scope := c.QueryParam("scope") // project or module
	if registry == "" {
		registry = "repo.maven.apache.org/maven2" // default to Maven Central
	}
	if coordinate == "" {
		return c.JSON(http.StatusBadRequest, "param coordinate is required")
	}
	if artifactVersion == "" {
		return c.JSON(http.StatusBadRequest, "param version is required")
	}
	if scope != "project" && scope != "module" {
		scope = "project"
	}
	if !util.IsValidMavenCoordinate(coordinate) {
		return c.JSON(http.StatusBadRequest, "query param coordinate is not a valid maven coordinate")
	}

	// lookup
	data, err := h.lookupService.LookupDependency(registry, coordinate)
	if err != nil {
		if errors.Is(err, service.ErrRepositoryNotFound) {
			return c.JSON(http.StatusOK, badge.NewDependencyBadge("repository not configured", badge.Error, theme))
		} else if errors.Is(err, service.ErrDependencyNotFound) {
			return c.JSON(http.StatusOK, badge.NewDependencyBadge("not configured", badge.Error, theme))
		}

		slog.Error("Error looking up dependency metadata", "err", err)
		return c.JSON(http.StatusInternalServerError, "internal server error")
	}

	// support "latest" as version
	if artifactVersion == "latest" {
		artifactVersion = data.Latest
	}

	// search version in data
	version, ok := data.Versions[artifactVersion]
	if !ok {
		return c.JSON(http.StatusOK, badge.NewDependencyBadge("pending verification", badge.Warning, theme))
	}

	// badge
	badgeText := fmt.Sprintf("%d/%d ok", version.FileStats.TotalReproducibleFiles, version.FileStats.TotalReproducibleFiles+version.FileStats.TotalNonReproducibleFiles)
	badgeStatus := util.Ternary(version.FileStats.TotalNonReproducibleFiles == 0, badge.Success, badge.Error)
	if scope == "module" {
		badgeText = fmt.Sprintf("%d/%d ok", version.FileStats.ModuleReproducibleFiles, version.FileStats.ModuleReproducibleFiles+version.FileStats.ModuleNonReproducibleFiles)
		badgeStatus = util.Ternary(version.FileStats.ModuleNonReproducibleFiles == 0, badge.Success, badge.Error)
	}

	return c.JSON(http.StatusOK, badge.NewDependencyBadge(
		badgeText,
		badgeStatus,
		theme),
	)
}
