package model

import (
	"encoding/xml"
)

type PomProject struct {
	XMLName              xml.Name                `xml:"project"`
	Packaging            string                  `xml:"packaging"`
	Dependencies         []PomDependency         `xml:"dependencies>dependency"`
	DependencyManagement PomDependencyManagement `xml:"dependencyManagement"`
}

type PomDependencyManagement struct {
	Dependencies []PomDependency `xml:"dependencies>dependency"`
}

type PomDependency struct {
	GroupId    string `xml:"groupId"`
	ArtifactId string `xml:"artifactId"`
	Version    string `xml:"version"`
	Scope      string `xml:"scope,omitempty"`
}
