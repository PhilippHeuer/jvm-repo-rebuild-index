package service

import (
	"errors"
	"fmt"
	"log/slog"
	"slices"

	"github.com/philippheuer/jvm-repo-rebuild-index/pkg/model"
	"github.com/philippheuer/jvm-repo-rebuild-index/pkg/util"
)

var registryNames = []string{
	"mavencentral",
	"gradlepluginportal",
}

var registryToNameMap = map[string]string{
	"repo.maven.apache.org/maven2": "mavencentral",
	"repo1.maven.org/maven2":       "mavencentral",
	"plugins.gradle.org/m2":        "gradlepluginportal",
}

var (
	ErrRepositoryNotFound = errors.New("repository is not supported")
	ErrDependencyNotFound = errors.New("dependency not found")
)

type DependencyLookupService interface {
	// FetchPom fetches the pom file for a given coordinate from the registry
	FetchPom(registry string, coordinate model.GAV) (*model.PomProject, error)
	// CollectCoordinates is a helper function that returns all dependency coordinates for bom artifacts, otherwise it returns the input coordinate
	CollectCoordinates(registry string, coordinate model.GAV) ([]model.GAV, error)
	LookupProject(registry string, coordinate model.GAV) (*model.Dependency, error)
	LookupDependency(registry string, coordinate model.GAV) (*model.Dependency, error)
	LookupDependencyVersion(registry string, coordinate model.GAV) (*model.Version, error)
}

type dependencyLookupService struct {
	LocalDir  string
	RemoteURL string
}

func NewDependencyLookupService(localDir, remoteURL string) DependencyLookupService {
	return &dependencyLookupService{
		LocalDir:  localDir,
		RemoteURL: remoteURL,
	}
}

func (s *dependencyLookupService) LookupProject(registry string, coordinate model.GAV) (*model.Dependency, error) {
	return s.lookup(registry, coordinate, "project")
}

func (s *dependencyLookupService) LookupDependency(registry string, coordinate model.GAV) (*model.Dependency, error) {
	return s.lookup(registry, coordinate, "maven")
}

func (s *dependencyLookupService) FetchPom(registry string, coordinate model.GAV) (*model.PomProject, error) {
	pom, err := util.LoadXMLFromURL[model.PomProject](fmt.Sprintf("https://%s/%s/%s-%s.pom", registry, coordinate.Path(false), coordinate.ArtifactId, coordinate.Version))
	if err != nil {
		return nil, errors.Join(ErrDependencyNotFound, err)
	}

	return &pom, nil
}

func (s *dependencyLookupService) CollectCoordinates(registry string, coordinate model.GAV) ([]model.GAV, error) {
	// collect coordinates
	var coordinates []model.GAV
	coordinates = append(coordinates, coordinate)

	pom, err := s.FetchPom(registry, coordinate)
	if err != nil {
		slog.Error("Error fetching pom", "err", err)
	}
	if pom.Packaging == "pom" {
		for _, dep := range pom.DependencyManagement.Dependencies {
			coordinates = append(coordinates, model.GAV{
				GroupId:    dep.GroupId,
				ArtifactId: dep.ArtifactId,
				Version:    dep.Version,
			})
		}
	}

	return coordinates, nil
}

func (s *dependencyLookupService) lookup(registry string, coordinate model.GAV, variant string) (*model.Dependency, error) {
	registry, err := toRegistryName(registry)
	if err != nil {
		return nil, err
	}

	// lookup via local filesystem
	if s.LocalDir != "" {
		data, err := util.LoadFromDisk[model.Dependency](fmt.Sprintf("%s/%s/%s/%s/index.json", s.LocalDir, registry, variant, coordinate.Path(true)))
		if err != nil {
			return nil, errors.Join(ErrDependencyNotFound, err)
		}

		return &data, nil
	}

	// lookup via remote url
	if s.RemoteURL != "" {
		data, err := util.LoadFromURL[model.Dependency](fmt.Sprintf("%s/%s/%s/%s/index.json", s.RemoteURL, registry, variant, coordinate.Path(true)))
		if err != nil {
			return nil, errors.Join(ErrDependencyNotFound, err)
		}

		return &data, nil
	}

	return nil, errors.New("no available method to lookup dependency metadata")
}

func (s *dependencyLookupService) LookupDependencyVersion(registry string, coordinate model.GAV) (*model.Version, error) {
	registry, err := toRegistryName(registry)
	if err != nil {
		return nil, err
	}

	// lookup via local filesystem
	if s.LocalDir != "" {
		data, err := util.LoadFromDisk[model.Version](fmt.Sprintf("%s/%s/maven/%s.json", s.LocalDir, registry, coordinate.Path(false)))
		if err != nil {
			return nil, errors.Join(ErrDependencyNotFound, err)
		}

		return &data, nil
	}

	// lookup via remote url
	if s.RemoteURL != "" {
		data, err := util.LoadFromURL[model.Version](fmt.Sprintf("%s/%s/maven/%s.json", s.RemoteURL, registry, coordinate.Path(false)))
		if err != nil {
			return nil, errors.Join(ErrDependencyNotFound, err)
		}

		return &data, nil
	}

	return nil, errors.New("no available method to lookup dependency metadata")
}

func toRegistryName(registryName string) (string, error) {
	if !slices.Contains(registryNames, registryName) {
		registryName = util.TrimURLProtocolAndTrailingSlash(registryName)

		newName, ok := registryToNameMap[registryName]
		if !ok {
			return "", ErrRepositoryNotFound
		}

		registryName = newName
	}

	return registryName, nil
}
