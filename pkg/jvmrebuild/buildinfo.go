package jvmrebuild

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/philippheuer/jvm-repo-rebuild-index/pkg/util"
)

type BuildInfo struct {
	SpecVersion string
	Name        string
	GroupID     string
	ArtifactID  string
	Version     string
	// Build instructions
	BuildTool string
	// Build environment
	JavaVersion string
	OSName      string
	// Source information
	SourceSCMUri string
	SourceSCMTag string
	// Output
	Outputs []Output
}

type Output struct {
	Coordinate string          `json:"coordinate"`
	Files      map[string]File `json:"files"`
}

type File struct {
	Size     string `json:"size,omitempty"`
	Checksum string `json:"checksum,omitempty"`
}

func ParseBuildInfo(file string) (BuildInfo, error) {
	kv, err := util.ParseFile(file)
	if err != nil {
		return BuildInfo{}, err
	}

	// kv
	buildInfo := BuildInfo{}
	if val, ok := kv["buildinfo.version"]; ok {
		buildInfo.SpecVersion = val
	}
	if val, ok := kv["name"]; ok {
		buildInfo.Name = val
	}
	if val, ok := kv["group-id"]; ok {
		buildInfo.GroupID = val
	}
	if val, ok := kv["artifact-id"]; ok {
		buildInfo.ArtifactID = val
	}
	if val, ok := kv["version"]; ok {
		buildInfo.Version = val
	}
	if val, ok := kv["build-tool"]; ok {
		buildInfo.BuildTool = val
	}
	if val, ok := kv["java.version"]; ok {
		buildInfo.JavaVersion = val
	}
	if val, ok := kv["os.name"]; ok {
		buildInfo.OSName = val
	}
	if val, ok := kv["source.scm.uri"]; ok {
		buildInfo.SourceSCMUri = val
	}
	if val, ok := kv["source.scm.tag"]; ok {
		buildInfo.SourceSCMTag = val
	}

	// outputs, there is some variance in the format
	if _, isVariant1 := kv["outputs.0.filename"]; isVariant1 {
		output := Output{
			Coordinate: fmt.Sprintf("%s:%s", buildInfo.GroupID, buildInfo.ArtifactID),
			Files:      make(map[string]File),
		}
		parseOutputFiles(kv, "outputs", &output)
		buildInfo.Outputs = append(buildInfo.Outputs, output)
	} else if _, isVariant2 := kv["outputs.0.coordinates"]; isVariant2 {
		for i := 0; ; i++ {
			coordinate, coordinateOk := kv[fmt.Sprintf("outputs.%d.coordinates", i)]
			if !coordinateOk {
				break
			}

			groupId := strings.SplitN(coordinate, ":", 2)[0]
			artifactId := strings.SplitN(coordinate, ":", 2)[1]
			if groupId == "" || artifactId == "" {
				slog.Warn("no group-id or artifact-id found for coordinate", "coordinate", coordinate)
				continue
			}

			output := Output{
				Coordinate: coordinate,
				Files:      make(map[string]File),
			}
			parseOutputFiles(kv, fmt.Sprintf("outputs.%d", i), &output)

			buildInfo.Outputs = append(buildInfo.Outputs, output)
		}
	} else if _, isVariant3 := kv["outputs.1.coordinates"]; isVariant3 {
		for i := 1; ; i++ {
			coordinate, coordinateOk := kv[fmt.Sprintf("outputs.%d.coordinates", i)]
			if !coordinateOk {
				break
			}

			groupId := strings.SplitN(coordinate, ":", 2)[0]
			artifactId := strings.SplitN(coordinate, ":", 2)[1]
			if groupId == "" || artifactId == "" {
				slog.Warn("no group-id or artifact-id found for coordinate", "coordinate", coordinate)
				continue
			}

			output := Output{
				Coordinate: coordinate,
				Files:      make(map[string]File),
			}
			parseOutputFiles(kv, fmt.Sprintf("outputs.%d", i), &output)
			buildInfo.Outputs = append(buildInfo.Outputs, output)
		}
	} else {
		return BuildInfo{}, fmt.Errorf("unsupported format")
	}

	return buildInfo, nil
}

func parseOutputFiles(kv map[string]string, prefix string, output *Output) {
	for i := 0; ; i++ {
		filename, filenameOk := kv[fmt.Sprintf("%s.%d.filename", prefix, i)]
		size, _ := kv[fmt.Sprintf("%s.%d.length", prefix, i)]
		checksum, _ := kv[fmt.Sprintf("%s.%d.checksums.sha512", prefix, i)]

		// stop if information is missing
		if !filenameOk {
			break
		}

		output.Files[filename] = File{
			Size:     size,
			Checksum: checksum,
		}
	}
}
