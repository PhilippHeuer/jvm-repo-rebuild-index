package main

import (
	"encoding/json"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/charlievieth/fastwalk"
)

type DependencyMetadata struct {
	GroupID    string                     `json:"group_id"`
	ArtifactID string                     `json:"artifact_id"`
	Versions   map[string]VersionMetadata `json:"versions"`
	Latest     VersionMetadata            `json:"latest"`
}

type VersionMetadata struct {
	Version              string   `json:"version"`
	DisplayName          string   `json:"display_name,omitempty"`
	SCMUri               string   `json:"scm_uri,omitempty"`
	SCMTag               string   `json:"scm_tag,omitempty"`
	BuildTool            string   `json:"build_tool,omitempty"`
	JavaVersion          string   `json:"java_version,omitempty"`
	OSName               string   `json:"os_name,omitempty"`
	Reproducible         bool     `json:"reproducible"`
	ReproducibleFiles    []string `json:"reproducible_files"`
	NotReproducibleFiles []string `json:"non_reproducible_files"`
}

func main() {
	inputDir := os.Args[1]
	outputDir := os.Args[2]
	slog.Info("generating index", "inputDir", inputDir, "outputDir", outputDir)

	// search for all maven-metadata.xml files in all subdirectories
	files, filesErr := findFiles(inputDir, "maven-metadata.xml")
	if filesErr != nil {
		slog.Error("failed to find maven-metadata.xml files", "error", filesErr)
		os.Exit(1)
	}

	allMetadata := make(map[string]DependencyMetadata)
	for _, mvnMetadataFile := range files {
		dependencyMetadata := DependencyMetadata{
			Versions: make(map[string]VersionMetadata),
		}
		dir := filepath.Dir(mvnMetadataFile)
		slog.Debug("found project", "path", mvnMetadataFile, "dir", filepath.Dir(mvnMetadataFile))

		buildInfoFiles, err := findFiles(dir, ".buildinfo")
		if err != nil {
			slog.Error("failed to find maven-metadata.xml files", "error", err)
			os.Exit(1)
		}

		slog.Debug("found buildinfo files", "count", len(buildInfoFiles))
		for _, buildInfoFile := range buildInfoFiles {
			slog.Debug("found buildinfo file", "path", buildInfoFile, "dir", filepath.Dir(buildInfoFile))

			buildInfo, err := parseFile(buildInfoFile)
			if err != nil {
				slog.Error("failed to parse buildinfo file", "error", err)
				continue
			}
			buildCompare, err := parseFile(strings.Replace(buildInfoFile, ".buildinfo", ".buildcompare", 1))
			if err != nil {
				slog.Error("failed to parse buildinfo file", "error", err)
				continue
			}

			buildInfoVersion := buildInfo["buildinfo.version"]
			if buildInfoVersion != "" && buildInfoVersion != "1.0-SNAPSHOT" {
				slog.Error("failed to find a supported buildinfo.version in buildinfo file", "file", buildInfoFile, "version", buildInfoVersion)
				continue
			}

			versionMetadata := VersionMetadata{
				Version:      buildCompare["version"],
				DisplayName:  buildInfo["name"],
				SCMUri:       buildInfo["source.scm.uri"],
				SCMTag:       buildInfo["source.scm.tag"],
				BuildTool:    buildInfo["build-tool"],
				JavaVersion:  buildInfo["java.version"],
				OSName:       buildInfo["os.name"],
				Reproducible: buildCompare["ko"] == "0" && buildCompare["ok"] != "0",
			}
			if rf, ok := buildCompare["okFiles"]; ok && rf == "" {
				versionMetadata.ReproducibleFiles = make([]string, 0)
			} else {
				versionMetadata.ReproducibleFiles = strings.Split(rf, " ")
			}
			if rf, ok := buildCompare["koFiles"]; ok && rf == "" {
				versionMetadata.NotReproducibleFiles = make([]string, 0)
			} else {
				versionMetadata.NotReproducibleFiles = strings.Split(rf, " ")
			}
			slog.Debug("parsed buildinfo data", "content", versionMetadata)

			dependencyMetadata.GroupID = buildInfo["group-id"]
			dependencyMetadata.ArtifactID = buildInfo["artifact-id"]
			dependencyMetadata.Versions[versionMetadata.Version] = versionMetadata

			// write version metadata to file
			writeErr := writeToFile(filepath.Join(outputDir, strings.ReplaceAll(dependencyMetadata.GroupID, ".", "/"), strings.ReplaceAll(dependencyMetadata.ArtifactID, ".", "/"), versionMetadata.Version+".json"), versionMetadata)
			if writeErr != nil {
				slog.Error("failed to write artifact metadata to file", "error", writeErr)
				os.Exit(1)
			}
		}

		// sort versions and find latest
		if len(dependencyMetadata.Versions) > 0 {
			var versions []*semver.Version
			for _, version := range dependencyMetadata.Versions {
				v, vErr := semver.NewVersion(version.Version)
				if vErr != nil {
					slog.Error("failed to parse version", "version", version.Version, "error", vErr)
					continue
				}

				versions = append(versions, v)
			}

			if len(versions) > 0 {
				sort.Sort(semver.Collection(versions))

				latestVersion := versions[len(versions)-1].String()
				dependencyMetadata.Latest = dependencyMetadata.Versions[latestVersion]
			}
		}

		// write artifact metadata to file
		writeErr := writeToFile(filepath.Join(outputDir, strings.ReplaceAll(dependencyMetadata.GroupID, ".", "/"), strings.ReplaceAll(dependencyMetadata.ArtifactID, ".", "/"), "index.json"), dependencyMetadata)
		if writeErr != nil {
			slog.Error("failed to write artifact metadata to file", "error", writeErr)
			os.Exit(1)
		}

		allMetadata[dependencyMetadata.GroupID+":"+dependencyMetadata.ArtifactID] = dependencyMetadata
	}

	// write all metadata to file
	writeErr := writeToFile(filepath.Join(outputDir, "index.json"), allMetadata)
	if writeErr != nil {
		slog.Error("failed to write artifact metadata to file", "error", writeErr)
		os.Exit(1)
	}
}

func findFiles(rootPath string, suffix string) ([]string, error) {
	var files []string

	conf := fastwalk.Config{
		Follow: false,
	}
	err := fastwalk.Walk(&conf, rootPath, func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if info.Name() == ".git" {
			return filepath.SkipDir
		}
		if info.IsDir() == false && strings.HasSuffix(filepath.Base(path), suffix) {
			files = append(files, path)
		}
		return nil
	})

	return files, err
}

func parseFile(filename string) (map[string]string, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return parseProperties(string(content)), nil
}

func parseProperties(content string) map[string]string {
	properties := make(map[string]string)

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
				value = value[1 : len(value)-1]
			}

			properties[key] = value
		}
	}

	return properties
}

func writeToFile(filename string, data any) error {
	if err := os.MkdirAll(filepath.Dir(filename), os.ModePerm); err != nil {
		return err
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}
