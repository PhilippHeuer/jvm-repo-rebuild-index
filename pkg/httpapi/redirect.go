package httpapi

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/philippheuer/jvm-repo-rebuild-index/pkg/model"
	"github.com/philippheuer/jvm-repo-rebuild-index/pkg/service"
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
	gav, err := model.NewGA(coordinate)
	if err != nil {
		return c.JSON(http.StatusBadRequest, "invalid maven coordinate")
	}

	// lookup
	data, err := h.lookupService.LookupDependency(registry, gav)
	if err != nil {
		if errors.Is(err, service.ErrRepositoryNotFound) {
			return c.JSON(http.StatusBadRequest, "repository not configured")
		} else if errors.Is(err, service.ErrDependencyNotFound) {
			return c.Redirect(http.StatusFound, "https://reproducible-builds.org/docs/jvm/") // redirect to documentation
		}

		slog.Error("Error looking up dependency metadata", "err", err)
		return c.JSON(http.StatusInternalServerError, "internal server error")
	}

	// redirect
	return c.Redirect(http.StatusFound, data.RebuildProjectUrl)
}
