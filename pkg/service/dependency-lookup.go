package service

import (
	"errors"
	"fmt"
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
	LookupProject(repository, coordinate string) (*model.DependencyMetadata, error)
	LookupDependency(repository, coordinate string) (*model.DependencyMetadata, error)
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

func (s *dependencyLookupService) LookupProject(registryName, coordinate string) (*model.DependencyMetadata, error) {
	return s.lookup(registryName, coordinate, "project")
}

func (s *dependencyLookupService) LookupDependency(registryName, coordinate string) (*model.DependencyMetadata, error) {
	return s.lookup(registryName, coordinate, "maven")
}

func (s *dependencyLookupService) lookup(registryName, coordinate, variant string) (*model.DependencyMetadata, error) {
	if !slices.Contains(registryNames, registryName) {
		registryName = util.TrimURLProtocolAndTrailingSlash(registryName)

		newName, ok := registryToNameMap[registryName]
		if !ok {
			return nil, ErrRepositoryNotFound
		}

		registryName = newName
	}

	// lookup via local filesystem
	if s.LocalDir != "" {
		data, err := util.LoadFromDisk[model.DependencyMetadata](fmt.Sprintf("%s/%s/%s/%s/index.json", s.LocalDir, registryName, variant, util.CoordinateToPath(coordinate, true)))
		if err != nil {
			return nil, errors.Join(ErrDependencyNotFound, err)
		}

		return &data, nil
	}

	// lookup via remote url
	if s.RemoteURL != "" {
		data, err := util.LoadFromURL[model.DependencyMetadata](fmt.Sprintf("%s/%s/%s/%s/index.json", s.RemoteURL, registryName, variant, util.CoordinateToPath(coordinate, true)))
		if err != nil {
			return nil, errors.Join(ErrDependencyNotFound, err)
		}

		return &data, nil
	}

	return nil, errors.New("no available method to lookup dependency metadata")
}
