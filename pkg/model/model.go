package model

type Repository struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type Project struct {
	RebuildProjectUrl string              `json:"rebuild_project_url,omitempty"`
	GroupID           string              `json:"group_id"`
	ArtifactID        string              `json:"artifact_id"`
	Modules           []string            `json:"modules,omitempty"`
	Versions          map[string]*Version `json:"versions"`
	Latest            string              `json:"latest"`
}

type Dependency struct {
	RebuildProjectUrl string              `json:"rebuild_project_url,omitempty"`
	GroupID           string              `json:"group_id"`
	ArtifactID        string              `json:"artifact_id"`
	Versions          map[string]*Version `json:"versions"`
	Latest            string              `json:"latest"`
}

type Version struct {
	Project          string          `json:"project,omitempty"`
	SCMUri           string          `json:"scm_uri,omitempty"`
	SCMTag           string          `json:"scm_tag,omitempty"`
	BuildTool        string          `json:"build_tool,omitempty"`
	BuildJavaVersion string          `json:"build_java_version,omitempty"`
	BuildOSName      string          `json:"build_os_name,omitempty"`
	Reproducible     bool            `json:"reproducible"`
	Files            map[string]File `json:"files,omitempty"`
	FileStats        FileStats       `json:"file_stats,omitempty"`
}

func (v *Version) SetTotalFileStats(allArtifacts map[string]File) {
	v.FileStats.TotalReproducibleFiles, v.FileStats.TotalNonReproducibleFiles = countReproducibleFiles(allArtifacts)
}

func (v *Version) SetModuleFileStats() {
	v.FileStats.ModuleReproducibleFiles, v.FileStats.ModuleNonReproducibleFiles = countReproducibleFiles(v.Files)
}

type FileStats struct {
	TotalReproducibleFiles     int `json:"total_reproducible"`
	TotalNonReproducibleFiles  int `json:"total_non_reproducible"`
	ModuleReproducibleFiles    int `json:"module_reproducible"`
	ModuleNonReproducibleFiles int `json:"module_non_reproducible"`
}

type File struct {
	Size         string `json:"size,omitempty"`
	Checksum     string `json:"checksum,omitempty"`
	Reproducible bool   `json:"reproducible"`
}

func countReproducibleFiles(files map[string]File) (reproducibleCount, nonReproducibleCount int) {
	for _, file := range files {
		if file.Reproducible {
			reproducibleCount++
		}
	}
	nonReproducibleCount = len(files) - reproducibleCount
	return reproducibleCount, nonReproducibleCount
}
