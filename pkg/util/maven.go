package util

import (
	"encoding/xml"
	"os"
	"strings"
	"unicode"
)

func IsValidMavenCoordinate(coordinate string) bool {
	for _, ch := range coordinate {
		if !unicode.IsLetter(ch) && !unicode.IsDigit(ch) && ch != '.' && ch != ':' && ch != '-' {
			return false
		}
	}
	return true
}

func CoordinateToPath(coordinate string, trimVersion bool) string {
	parts := strings.SplitN(coordinate, ":", 3)
	groupAndArtifact := strings.NewReplacer(".", "/", ":", "/").Replace(parts[0] + ":" + parts[1])

	if trimVersion || len(parts) < 3 {
		return groupAndArtifact
	}
	return groupAndArtifact + "/" + parts[2]
}

type MavenMetadata struct {
	GroupID    string     `xml:"groupId"`
	ArtifactID string     `xml:"artifactId"`
	Versioning Versioning `xml:"versioning"`
}

type Versioning struct {
	Latest      string   `xml:"latest"`
	Release     string   `xml:"release"`
	Versions    []string `xml:"versions>version"`
	LastUpdated string   `xml:"lastUpdated"`
}

func ParseMavenMetadataFile(filename string) (MavenMetadata, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return MavenMetadata{}, err
	}

	return ParseMavenMetadata(content)
}

func ParseMavenMetadata(content []byte) (MavenMetadata, error) {
	var metadata MavenMetadata
	err := xml.Unmarshal([]byte(content), &metadata)
	if err != nil {
		return MavenMetadata{}, err
	}
	return metadata, nil
}
