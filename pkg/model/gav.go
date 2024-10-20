package model

import (
	"errors"
	"strings"
	"unicode"

	"github.com/philippheuer/jvm-repo-rebuild-index/pkg/util"
)

type GAV struct {
	GroupId    string `json:"groupId"`
	ArtifactId string `json:"artifactId"`
	Version    string `json:"version,omitempty"`
}

func (gav *GAV) Coordinate() string {
	if gav.Version == "" {
		return gav.GroupId + ":" + gav.ArtifactId
	}
	return gav.GroupId + ":" + gav.ArtifactId + ":" + gav.Version
}

func (gav *GAV) Path(trimVersion bool) string {
	return util.CoordinateToPath(gav.Coordinate(), trimVersion)
}

// NewGAV creates a new GAV (groupId, artifactId, version) struct from a Maven coordinate
func NewGAV(coordinate string) (GAV, error) {
	if !IsValidMavenCoordinate(coordinate) {
		return GAV{}, errors.New("invalid Maven coordinate: contains illegal characters")
	}

	parts := strings.Split(coordinate, ":")
	if len(parts) != 3 {
		return GAV{}, errors.New("invalid Maven coordinate: expected format is 'groupId:artifactId:version'")
	}

	return GAV{
		GroupId:    parts[0],
		ArtifactId: parts[1],
		Version:    parts[2],
	}, nil
}

// NewGA creates a new GAV (groupId, artifactId) struct from a Maven coordinate
func NewGA(coordinate string) (GAV, error) {
	if !IsValidMavenCoordinate(coordinate) {
		return GAV{}, errors.New("invalid Maven coordinate: contains illegal characters")
	}

	parts := strings.Split(coordinate, ":")
	if len(parts) != 2 {
		return GAV{}, errors.New("invalid Maven coordinate: expected format is 'groupId:artifactId'")
	}

	return GAV{
		GroupId:    parts[0],
		ArtifactId: parts[1],
	}, nil
}

func IsValidMavenCoordinate(coordinate string) bool {
	for _, ch := range coordinate {
		if !unicode.IsLetter(ch) && !unicode.IsDigit(ch) && ch != '.' && ch != ':' && ch != '-' {
			return false
		}
	}
	return true
}
