package model

type DependencyMetadata struct {
	GroupID    string                     `json:"group_id"`
	ArtifactID string                     `json:"artifact_id"`
	Modules    []string                   `json:"modules,omitempty"`
	Versions   map[string]VersionMetadata `json:"versions"`
	Latest     string                     `json:"latest"`
}

type VersionMetadata struct {
	Project              string              `json:"project,omitempty"`
	SCMUri               string              `json:"scm_uri,omitempty"`
	SCMTag               string              `json:"scm_tag,omitempty"`
	BuildTool            string              `json:"build_tool,omitempty"`
	BuildJavaVersion     string              `json:"build_java_version,omitempty"`
	BuildOSName          string              `json:"build_os_name,omitempty"`
	Reproducible         bool                `json:"reproducible"`
	ReproducibleFiles    int                 `json:"reproducible_files"`
	NonReproducibleFiles int                 `json:"non_reproducible_files"`
	Artifacts            map[string]Artifact `json:"artifacts,omitempty"`
}

func (v *VersionMetadata) SetReproducibleFilesByArtifacts() {
	reproducibleCount := 0
	for _, artifact := range v.Artifacts {
		if artifact.Reproducible {
			reproducibleCount++
		}
	}

	v.ReproducibleFiles = reproducibleCount
	v.NonReproducibleFiles = len(v.Artifacts) - reproducibleCount
}

type Artifact struct {
	Size         string `json:"size,omitempty"`
	Checksum     string `json:"checksum,omitempty"`
	Reproducible bool   `json:"reproducible"`
}

type Badge struct {
	SchemaVersion int    `json:"schemaVersion"`
	Label         string `json:"label"`
	Message       string `json:"message"`
	Color         string `json:"color,omitempty"`
	LabelColor    string `json:"labelColor"`
	IsError       bool   `json:"isError"`
	Style         string `json:"style"`
}
