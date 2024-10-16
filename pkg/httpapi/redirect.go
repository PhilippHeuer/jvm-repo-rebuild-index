package httpapi

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/philippheuer/jvm-repo-rebuild-index/pkg/service"
	"github.com/philippheuer/jvm-repo-rebuild-index/pkg/util"
)

func (h handlers) redirectHandler(c echo.Context) error {
	registry := c.Param("registry")
	coordinate := c.Param("coordinate")
	if registry == "" {
		registry = "repo.maven.apache.org/maven2" // default to Maven Central
	}
	if coordinate == "" {
		return c.JSON(http.StatusBadRequest, "query param coordinate is required")
	}
	if !util.IsValidMavenCoordinate(coordinate) {
		return c.JSON(http.StatusBadRequest, "query param coordinate is not a valid maven coordinate")
	}

	// lookup
	data, err := h.lookupService.LookupDependency(registry, coordinate)
	if err != nil {
		if errors.Is(err, service.ErrRepositoryNotFound) {
			return c.JSON(http.StatusBadRequest, "repository not configured")
		} else if errors.Is(err, service.ErrDependencyNotFound) {
			return c.JSON(http.StatusBadRequest, "artifact not configured")
		}

		slog.Error("Error looking up dependency metadata", "err", err)
		return c.JSON(http.StatusInternalServerError, "internal server error")
	}

	// redirect
	return c.Redirect(http.StatusFound, data.RebuildProjectUrl)
}
