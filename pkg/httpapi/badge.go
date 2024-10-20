package httpapi

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"slices"

	"github.com/labstack/echo/v4"
	"github.com/philippheuer/jvm-repo-rebuild-index/pkg/badge"
	"github.com/philippheuer/jvm-repo-rebuild-index/pkg/model"
	"github.com/philippheuer/jvm-repo-rebuild-index/pkg/service"
	"github.com/philippheuer/jvm-repo-rebuild-index/pkg/sonatype"
	"github.com/philippheuer/jvm-repo-rebuild-index/pkg/util"
)

func (h handlers) projectBadgeHandler(c echo.Context) error {
	registry := c.Param("registry")
	coordinate := c.Param("coordinate")
	artifactVersion := c.Param("version")
	theme := c.QueryParam("theme")

	registry, err := url.QueryUnescape(registry)
	if err != nil {
		return c.JSON(http.StatusBadRequest, "failed to decode registry")
	}
	coordinate, err = url.QueryUnescape(coordinate)
	if err != nil {
		return c.JSON(http.StatusBadRequest, "failed to decode coordinate")
	}

	if registry == "" {
		registry = "repo.maven.apache.org/maven2" // default to Maven Central
	}
	if coordinate == "" {
		return c.JSON(http.StatusBadRequest, "param coordinate is required")
	}
	if artifactVersion == "" {
		return c.JSON(http.StatusBadRequest, "param version is required")
	}
	gav, err := model.NewGAV(coordinate + ":" + artifactVersion)
	if err != nil {
		return c.JSON(http.StatusBadRequest, "invalid maven coordinate")
	}

	// lookup
	data, err := h.lookupService.LookupProject(registry, gav)
	if err != nil {
		if errors.Is(err, service.ErrRegistryNotFound) {
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

	registry, err := url.QueryUnescape(registry)
	if err != nil {
		return c.JSON(http.StatusBadRequest, "failed to decode registry")
	}
	coordinate, err = url.QueryUnescape(coordinate)
	if err != nil {
		return c.JSON(http.StatusBadRequest, "failed to decode coordinate")
	}

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
	gav, err := model.NewGAV(coordinate + ":" + artifactVersion)
	if err != nil {
		return c.JSON(http.StatusBadRequest, "invalid maven coordinate")
	}

	// lookup
	data, err := h.lookupService.LookupDependency(registry, gav)
	if err != nil {
		if errors.Is(err, service.ErrRegistryNotFound) {
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

func (h handlers) transitiveDependencyBadgeHandler(c echo.Context) error {
	registry := c.Param("registry")
	coordinate := c.Param("coordinate")
	artifactVersion := c.Param("version")
	theme := c.QueryParam("theme")

	registry, err := url.QueryUnescape(registry)
	if err != nil {
		return c.JSON(http.StatusBadRequest, "failed to decode registry")
	}
	coordinate, err = url.QueryUnescape(coordinate)
	if err != nil {
		return c.JSON(http.StatusBadRequest, "failed to decode coordinate")
	}

	if registry == "" {
		registry = "repo.maven.apache.org/maven2" // default to Maven Central
	}
	if coordinate == "" {
		return c.JSON(http.StatusBadRequest, "param coordinate is required")
	}
	if artifactVersion == "" {
		return c.JSON(http.StatusBadRequest, "param version is required")
	}
	gav, err := model.NewGAV(coordinate + ":" + artifactVersion)
	if err != nil {
		return c.JSON(http.StatusBadRequest, "invalid maven coordinate")
	}

	// collect coordinates (includes special handling for BOMs)
	coordinates, err := h.lookupService.CollectCoordinates(registry, gav)
	if err != nil {
		slog.Error("Error collecting coordinates", "err", err)
		return c.JSON(http.StatusInternalServerError, "internal server error")
	}

	// lookup transitive dependencies
	var dependencies []sonatype.Component
	for _, cord := range coordinates {
		dep, dErr := sonatype.FetchAllDependencies(fmt.Sprintf("pkg:maven/%s/%s@%s", cord.GroupId, cord.ArtifactId, cord.Version))
		if dErr != nil {
			slog.Error("Error fetching transitive dependencies", "err", err)
			return c.JSON(http.StatusInternalServerError, "internal server error")
		}

		dependencies = append(dependencies, dep...)
	}
	dependencies = slices.CompactFunc(dependencies, sonatype.ComponentEquals)

	// evaluate dependencies
	var allDependencies []string
	var reproducibleDependencies []string
	for _, dep := range dependencies {
		allDependencies = append(allDependencies, fmt.Sprintf("%s:%s:%s", dep.DependencyNamespace, dep.DependencyName, dep.DependencyVersion))

		dResult, dErr := h.lookupService.LookupDependencyVersion(registry, model.GAV{GroupId: dep.DependencyNamespace, ArtifactId: dep.DependencyName, Version: dep.DependencyVersion})
		if dErr != nil {
			continue
		}

		if dResult.FileStats.TotalNonReproducibleFiles == 0 {
			reproducibleDependencies = append(reproducibleDependencies, fmt.Sprintf("%s:%s:%s", dep.DependencyNamespace, dep.DependencyName, dep.DependencyVersion))
		}
	}

	// badge
	badgeText := fmt.Sprintf("%d/%d dep(s)", len(reproducibleDependencies), len(allDependencies))
	badgeStatus := util.Ternary(len(reproducibleDependencies) == len(allDependencies), badge.Success, badge.Warning)
	return c.JSON(http.StatusOK, badge.NewDependencyBadge(
		badgeText,
		badgeStatus,
		theme),
	)
}
