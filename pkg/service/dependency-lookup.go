package service

import (
	"errors"
	"fmt"

	"github.com/philippheuer/reproducible-central-index/pkg/model"
	"github.com/philippheuer/reproducible-central-index/pkg/util"
)

var repositoryToNameMap = map[string]string{
	"repo.maven.apache.org/maven2": "mavencentral",
	"repo1.maven.org/maven2":       "mavencentral",
}

var (
	ErrRepositoryNotFound = errors.New("repository is not supported")
	ErrDependencyNotFound = errors.New("dependency not found")
)

type DependencyLookupService interface {
	LookupDependencyMetadata(repository, coordinate string) (*model.DependencyMetadata, error)
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

func (s *dependencyLookupService) LookupDependencyMetadata(repository, coordinate string) (*model.DependencyMetadata, error) {
	// lookup repository name
	repositoryName, ok := repositoryToNameMap[repository]
	if !ok {
		return nil, ErrRepositoryNotFound
	}

	// lookup dependency
	data, err := util.LoadFromURL[model.DependencyMetadata](fmt.Sprintf("%s/%s/maven/%s/index.json", s.RemoteURL, repositoryName, util.CoordinateToPath(coordinate, true)))
	if err != nil {
		return nil, errors.Join(ErrDependencyNotFound, err)
	}

	return &data, nil
}
