package httpapi

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/philippheuer/jvm-repo-rebuild-index/pkg/service"
)

type handlers struct {
	lookupService service.DependencyLookupService
}

var ErrStartingServer = errors.New("error starting server")

func Serve(port int, indexDir string, indexURL string) error {
	// config
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	// middlewares
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// services
	handlerStruct := handlers{
		lookupService: service.NewDependencyLookupService(indexDir, indexURL),
	}

	// handlers
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, "OK")
	})

	e.GET("/v1/badge/reproducible/maven/:coordinate/:version", handlerStruct.dependencyBadgeHandler)
	e.GET("/v1/badge/reproducible/maven/:registry/:coordinate/:version", handlerStruct.dependencyBadgeHandler)

	e.GET("/v1/badge/reproducible-dependencies/maven/:registry/:coordinate/:version", handlerStruct.transitiveDependencyBadgeHandler)
	e.GET("/v1/badge/reproducible-dependencies/maven/:coordinate/:version", handlerStruct.transitiveDependencyBadgeHandler)

	e.GET("/v1/badge/reproducible/project/:coordinate/:version", handlerStruct.projectBadgeHandler)
	e.GET("/v1/badge/reproducible/project/:registry/:coordinate/:version", handlerStruct.projectBadgeHandler)

	e.GET("/v1/redirect/reproducible/maven/:coordinate/:version", handlerStruct.redirectHandler)
	e.GET("/v1/redirect/reproducible/maven/:registry/:coordinate/:version", handlerStruct.redirectHandler)

	// start
	startErr := e.Start(fmt.Sprintf(":%d", port))
	if startErr != nil {
		if startErr.Error() == "http: Server closed" {
			return nil
		}

		return errors.Join(ErrStartingServer, startErr)
	}

	return nil
}
