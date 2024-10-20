package util

import (
	"encoding/xml"
	"os"
)

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
	err := xml.Unmarshal(content, &metadata)
	if err != nil {
		return MavenMetadata{}, err
	}
	return metadata, nil
}
