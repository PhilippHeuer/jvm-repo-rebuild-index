package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"sync"

	"github.com/Masterminds/semver/v3"
	"github.com/philippheuer/reproducible-central-index/pkg/util"
)

type DependencyMetadata struct {
	GroupID    string                     `json:"group_id"`
	ArtifactID string                     `json:"artifact_id"`
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

func main() {
	if len(os.Args) != 3 {
		slog.Error("usage: <inputDir> <outputDir>")
		os.Exit(1)
	}
	inputDir := os.Args[1]
	outputDir := os.Args[2]
	slog.Info("generating index", "inputDir", inputDir, "outputDir", outputDir)

	// search for all maven-metadata.xml files in all subdirectories
	files, filesErr := util.FindFiles(inputDir, "maven-metadata.xml")
	if filesErr != nil {
		slog.Error("failed to find maven-metadata.xml files", "error", filesErr)
		os.Exit(1)
	}

	// process all files concurrently
	allMetadata := make(map[string]*DependencyMetadata)
	var wg sync.WaitGroup
	var mu sync.Mutex
	for _, mvnMetadataFile := range files {
		wg.Add(1)

		go func(file string) {
			defer wg.Done()
			data, err := processFile(file, outputDir)
			if err != nil {
				slog.Error("failed to process file", "error", err)
				return
			}

			// safely add metadata to map
			mu.Lock()
			for k, v := range data { // new
				if !strings.HasPrefix(k, "io.github.xanthic") {
					//continue
				}

				allMetadata[k] = v
				slog.Warn("added metadata", "key", k, "versions", len(v.Versions))
			}
			mu.Unlock()
		}(mvnMetadataFile)
	}
	wg.Wait()
	slog.Info("generated index", "count", len(allMetadata))

	// write data to filesystem
	var wmg sync.WaitGroup
	for _, v := range allMetadata {
		wmg.Add(1)

		go func(data *DependencyMetadata) {
			defer wmg.Done()
			// write artifact data
			writeErr := util.WriteToFile(filepath.Join(outputDir, strings.ReplaceAll(data.GroupID, ".", "/"), strings.ReplaceAll(data.ArtifactID, ".", "/"), "index.json"), v)
			if writeErr != nil {
				slog.Error("failed to write artifact metadata to file", "error", writeErr)
				os.Exit(1)
			}

			// write version data
			for version, versionMetadata := range data.Versions {
				writeVerErr := util.WriteToFile(filepath.Join(outputDir, strings.ReplaceAll(data.GroupID, ".", "/"), strings.ReplaceAll(data.ArtifactID, ".", "/"), version+".json"), versionMetadata)
				if writeVerErr != nil {
					slog.Error("failed to write artifact metadata to file", "error", writeVerErr)
					os.Exit(1)
				}
			}

			// write badge data
			if data.Latest != "" {
				latestVersion, latestVersionFound := data.Versions[data.Latest]
				if latestVersionFound {
					badge := Badge{
						SchemaVersion: 1,
						Label:         "Reproducible Builds",
						LabelColor:    "1e5b96",
						Color:         util.Ternary(latestVersion.Reproducible, "ok", "error"),
						Message:       fmt.Sprintf("%s - [%d ok / %d error]", data.Latest, latestVersion.ReproducibleFiles, latestVersion.NonReproducibleFiles),
						IsError:       latestVersion.Reproducible == false,
						Style:         "flat",
					}
					writeErr = util.WriteToFile(filepath.Join(outputDir, strings.ReplaceAll(data.GroupID, ".", "/"), strings.ReplaceAll(data.ArtifactID, ".", "/"), "badge.json"), badge)
					if writeErr != nil {
						slog.Error("failed to write artifact metadata to file", "error", writeErr)
						os.Exit(1)
					}
				}
			}
		}(v)
	}
	wmg.Wait()

	// write all metadata to file (disabled for now, this could very quickly use up the available github-pages bandwidth)
	/*
		writeErr := util.WriteToFile(filepath.Join(outputDir, "index.json"), allMetadata)
		if writeErr != nil {
			slog.Error("failed to write artifact metadata to file", "error", writeErr)
			os.Exit(1)
		}
	*/
}

func processFile(mvnMetadataFile string, outputDir string) (map[string]*DependencyMetadata, error) {
	result := make(map[string]*DependencyMetadata)

	dir := filepath.Dir(mvnMetadataFile)
	slog.Debug("found project", "path", mvnMetadataFile, "dir", filepath.Dir(mvnMetadataFile))

	buildInfoFiles, err := util.FindFiles(dir, ".buildinfo")
	if err != nil {
		slog.Error("failed to find maven-metadata.xml files", "error", err)
		return nil, errors.Join(errors.New("failed to find buildinfo files"), err)
	}

	slog.Debug("found buildinfo files", "count", len(buildInfoFiles))
	for _, buildInfoFile := range buildInfoFiles {
		slog.Debug("found buildinfo file", "path", buildInfoFile, "dir", filepath.Dir(buildInfoFile))

		buildInfo, err := util.ParseFile(buildInfoFile)
		if err != nil {
			slog.Error("failed to parse buildinfo file", "error", err)
			continue
		}
		buildCompare, err := util.ParseFile(strings.Replace(buildInfoFile, ".buildinfo", ".buildcompare", 1))
		if err != nil {
			slog.Error("failed to parse buildinfo file", "error", err)
			continue
		}

		buildInfoVersion := buildInfo["buildinfo.version"]
		if buildInfoVersion != "" && buildInfoVersion != "1.0-SNAPSHOT" {
			slog.Error("failed to find a supported buildinfo.version in buildinfo file", "file", buildInfoFile, "version", buildInfoVersion)
			continue
		}
		groupId := buildInfo["group-id"]
		artifactId := buildInfo["artifact-id"]
		version := buildCompare["version"]

		var reproducibleFiles []string
		if rf, ok := buildCompare["okFiles"]; ok {
			reproducibleFiles = strings.Split(rf, " ")
		}
		slog.Debug("parsed buildinfo and buildcompare file", "file", buildInfoFile, "version", version)

		// iterate over all outputs (look for key matching e.g. outputs.3.coordinates in buildInfo)
		for key, coordinate := range buildInfo {
			if !strings.HasPrefix(key, "outputs.") || !strings.HasSuffix(key, ".coordinates") {
				continue // skip, if key is not a coordinate key
			}

			versionMetadata := VersionMetadata{
				Project:          buildInfo["name"],
				SCMUri:           buildInfo["source.scm.uri"],
				SCMTag:           buildInfo["source.scm.tag"],
				BuildTool:        buildInfo["build-tool"],
				BuildJavaVersion: buildInfo["java.version"],
				BuildOSName:      buildInfo["os.name"],
				Reproducible:     buildCompare["ko"] == "0" && buildCompare["ok"] != "0",
				Artifacts:        make(map[string]Artifact),
			}
			keyPrefix := strings.TrimSuffix(key, ".coordinates")
			slog.Warn("found artifact", "key", key, "coordinate", coordinate)

			groupId = strings.SplitN(coordinate, ":", 2)[0]
			artifactId = strings.SplitN(coordinate, ":", 2)[1]
			if groupId == "" || artifactId == "" {
				slog.Warn("no group-id or artifact-id found for version", "version", version, "coordinate", coordinate)
				continue
			}

			for i := 0; ; i++ {
				filename, filenameOk := buildInfo[fmt.Sprintf("%s.%d.filename", keyPrefix, i)]
				size, sizeOK := buildInfo[fmt.Sprintf("%s.%d.length", keyPrefix, i)]
				checksum, checksumOk := buildInfo[fmt.Sprintf("%s.%d.checksums.sha512", keyPrefix, i)]

				// stop, if information is missing
				if !filenameOk || !sizeOK || !checksumOk {
					break
				}

				// only add artifacts that match the artifactId
				if strings.HasPrefix(filename, artifactId+"-"+version) {
					reproducible := slices.Contains(reproducibleFiles, filename)
					versionMetadata.Artifacts[filename] = Artifact{
						Size:         size,
						Checksum:     checksum,
						Reproducible: reproducible,
					}

				}
			}

			// count reproducible artifacts
			reproducibleCount := 0
			for _, artifact := range versionMetadata.Artifacts {
				if artifact.Reproducible {
					reproducibleCount++
				}
			}
			versionMetadata.ReproducibleFiles = reproducibleCount
			versionMetadata.NonReproducibleFiles = len(versionMetadata.Artifacts) - reproducibleCount

			// create or append to result
			if _, ok := result[coordinate]; !ok {
				result[coordinate] = &DependencyMetadata{
					GroupID:    groupId,
					ArtifactID: artifactId,
					Versions:   map[string]VersionMetadata{version: versionMetadata},
				}
			} else {
				result[coordinate].Versions[version] = versionMetadata
			}
		}
	}

	// find latest version
	for k, v := range result {
		var versions []*semver.Version
		for ver, _ := range v.Versions {
			version, err := semver.NewVersion(ver)
			if err != nil {
				slog.Error("failed to parse version", "version", ver, "error", err)
				continue
			}
			versions = append(versions, version)
		}
		if len(versions) > 0 {
			sort.Sort(semver.Collection(versions))
			latestVersion := versions[len(versions)-1].String()
			result[k].Latest = latestVersion
		}
	}

	return result, nil
}
