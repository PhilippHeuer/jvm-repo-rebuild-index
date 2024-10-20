package model

import (
	"errors"
	"strings"
	"unicode"
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
	groupAndArtifact := strings.NewReplacer(".", "/", ":", "/").Replace(gav.GroupId + ":" + gav.ArtifactId)

	if trimVersion || gav.Version == "" {
		return groupAndArtifact
	}
	return groupAndArtifact + "/" + gav.Version
}

// NewGAV creates a new GAV (groupId, artifactId, version) struct from a Maven coordinate
func NewGAV(coordinate string) (GAV, error) {
	return parseMavenCoordinate(coordinate)
}

func NewGAVIgnoreError(coordinate string) GAV {
	gav, err := parseMavenCoordinate(coordinate)
	if err != nil {
		return GAV{}
	}
	return gav
}

// parseMavenCoordinate is a helper function that parses Maven coordinates into a struct
func parseMavenCoordinate(coordinate string) (GAV, error) {
	if !isValidMavenCoordinate(coordinate) {
		return GAV{}, errors.New("invalid Maven coordinate: contains illegal characters")
	}

	parts := strings.Split(coordinate, ":")
	if len(parts) < 2 || len(parts) > 3 {
		return GAV{}, errors.New("invalid Maven coordinate: expected format is 'groupId:artifactId' or 'groupId:artifactId:version'")
	}

	gav := GAV{
		GroupId:    parts[0],
		ArtifactId: parts[1],
	}
	if len(parts) == 3 {
		gav.Version = parts[2]
	}

	return gav, nil
}

func isValidMavenCoordinate(coordinate string) bool {
	for _, ch := range coordinate {
		if !unicode.IsLetter(ch) && !unicode.IsDigit(ch) && ch != '.' && ch != ':' && ch != '-' {
			return false
		}
	}
	return true
}
