package cmd

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
	"github.com/philippheuer/reproducible-central-index/pkg/model"
	"github.com/philippheuer/reproducible-central-index/pkg/util"
	"github.com/spf13/cobra"
)

const MaxConcurrency = 250

func indexCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "index",
		Short: "generate index files",
		Run: func(cmd *cobra.Command, args []string) {
			inputDir, _ := cmd.Flags().GetString("input")
			outputDir, _ := cmd.Flags().GetString("output")
			if inputDir == "" || outputDir == "" {
				slog.Error("input and output directory are required")
				os.Exit(1)
			}
			slog.Info("generating index", "inputDir", inputDir, "outputDir", outputDir)

			// search for maven-metadata.xml across all directories
			files, filesErr := util.FindFiles(inputDir, "maven-metadata.xml")
			if filesErr != nil {
				slog.Error("failed to find maven-metadata.xml files", "error", filesErr)
				os.Exit(1)
			}

			// process all files concurrently
			depMetadata, projectMetadata := processFiles(files)
			slog.Info("generated index", "projects", len(projectMetadata), "artifacts", len(depMetadata))

			// write data to filesystem
			writeProjectIndexToFilesystem(outputDir, projectMetadata)
			writeDependencyIndexToFilesystem(outputDir, depMetadata)

			// write all metadata to file (disabled for now, this could very quickly use up the available github-pages bandwidth)
			/*
				writeErr := util.WriteToFile(filepath.Join(outputDir, "index.json"), allMetadata)
				if writeErr != nil {
					slog.Error("failed to write artifact metadata to file", "error", writeErr)
					os.Exit(1)
				}
			*/
		},
	}

	cmd.Flags().StringP("input", "i", "", "Input Directory")
	cmd.Flags().StringP("output", "o", "", "Output Directory")

	return cmd
}

func processFiles(files []string) (map[string]*model.DependencyMetadata, map[string]*model.ProjectMetadata) {
	depMetadata := make(map[string]*model.DependencyMetadata)
	projectMetadata := make(map[string]*model.ProjectMetadata)
	var wg sync.WaitGroup
	var mu sync.Mutex
	sem := make(chan struct{}, MaxConcurrency) // semaphore to limit concurrency
	for _, mvnMetadataFile := range files {
		wg.Add(1)
		sem <- struct{}{} // acquire semaphore

		go func(file string) {
			defer wg.Done()
			defer func() {
				<-sem // release semaphore
			}()

			data, projectData, err := processFile(file)
			if err != nil {
				slog.Error("failed to process file", "error", err)
				return
			}

			// safely add metadata to map
			mu.Lock()
			for k, v := range data {
				depMetadata[k] = v
				slog.Debug("added to dependency data", "key", k, "versions", len(v.Versions))
			}
			for k, v := range projectData {
				projectMetadata[k] = v
				slog.Debug("added to project data", "key", k, "versions", len(v.Versions))
			}
			mu.Unlock()
		}(mvnMetadataFile)
	}
	wg.Wait()

	return depMetadata, projectMetadata
}

func processFile(mvnMetadataFile string) (map[string]*model.DependencyMetadata, map[string]*model.ProjectMetadata, error) {
	result := make(map[string]*model.DependencyMetadata)     // individual artifact metadata
	projectResult := make(map[string]*model.ProjectMetadata) // project metadata

	dir := filepath.Dir(mvnMetadataFile)
	slog.Debug("found project", "path", mvnMetadataFile, "dir", filepath.Dir(mvnMetadataFile))

	// TODO: this will generate invalid links when using with other repos than reproducible-central
	contentIndex := strings.Index(dir, "/content")
	overviewUrl := ""
	if contentIndex != -1 {
		suffix := dir[contentIndex:]
		overviewUrl = "https://github.com/jvm-repo-rebuild/reproducible-central/blob/master" + suffix + "/README.md"
	}

	buildInfoFiles, err := util.FindFiles(dir, ".buildinfo")
	if err != nil {
		return nil, nil, errors.Join(errors.New("failed to find buildinfo files"), err)
	}

	slog.Debug("found buildinfo files", "count", len(buildInfoFiles))
	for _, buildInfoFile := range buildInfoFiles {
		slog.Debug("found buildinfo file", "path", buildInfoFile, "dir", filepath.Dir(buildInfoFile))

		buildInfo, buildInfoErr := util.ParseFile(buildInfoFile)
		if buildInfoErr != nil {
			slog.Error("failed to parse buildinfo file", "error", buildInfoErr)
			continue
		}
		buildCompare, buildCompareErr := util.ParseFile(strings.Replace(buildInfoFile, ".buildinfo", ".buildcompare", 1))
		if buildCompareErr != nil {
			slog.Error("failed to parse buildinfo file", "error", buildCompareErr)
			continue
		}

		buildInfoVersion := buildInfo["buildinfo.version"]
		if buildInfoVersion != "" && buildInfoVersion != "1.0-SNAPSHOT" {
			slog.Error("failed to find a supported buildinfo.version in buildinfo file", "file", buildInfoFile, "version", buildInfoVersion)
			continue
		}
		projectGroupId := buildInfo["group-id"]
		projectArtifactId := buildInfo["artifact-id"]
		version := buildCompare["version"]

		versionData := model.VersionMetadata{
			Project:          buildInfo["name"],
			SCMUri:           buildInfo["source.scm.uri"],
			SCMTag:           buildInfo["source.scm.tag"],
			BuildTool:        buildInfo["build-tool"],
			BuildJavaVersion: buildInfo["java.version"],
			BuildOSName:      buildInfo["os.name"],
			Reproducible:     buildCompare["ko"] == "0" && buildCompare["ok"] != "0",
		}
		allArtifacts := make(map[string]model.Artifact)
		var allCoordinates []string

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

			vd := versionData
			vd.Artifacts = make(map[string]model.Artifact)
			keyPrefix := strings.TrimSuffix(key, ".coordinates")
			slog.Debug("found artifact", "key", key, "coordinate", coordinate)

			groupId := strings.SplitN(coordinate, ":", 2)[0]
			artifactId := strings.SplitN(coordinate, ":", 2)[1]
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
					vd.Artifacts[filename] = model.Artifact{
						Size:         size,
						Checksum:     checksum,
						Reproducible: reproducible,
					}
					allArtifacts[filename] = vd.Artifacts[filename]
				}
			}
			vd.SetReproducibleFilesByArtifacts()

			// create or append to result
			if _, ok := result[coordinate]; !ok {
				result[coordinate] = &model.DependencyMetadata{
					ReproducibleOverviewURL: overviewUrl,
					GroupID:                 groupId,
					ArtifactID:              artifactId,
					Versions:                map[string]*model.VersionMetadata{version: &vd},
				}
				allCoordinates = append(allCoordinates, coordinate)
			} else {
				result[coordinate].Versions[version] = &vd
			}
		}

		// set reproducible file count for the entire project
		for rk := range result {
			if result[rk].Versions[version] == nil {
				continue
			}

			result[rk].Versions[version].SetProjectReproducibleFilesByArtifacts(allArtifacts)
		}

		// append project metadata
		versionData.Artifacts = allArtifacts
		versionData.SetReproducibleFilesByArtifacts()
		projectKey := projectGroupId + ":" + projectArtifactId
		if _, ok := projectResult[projectKey]; !ok {
			projectResult[projectKey] = &model.ProjectMetadata{
				ReproducibleOverviewURL: overviewUrl,
				GroupID:                 projectGroupId,
				ArtifactID:              projectArtifactId,
				Modules:                 allCoordinates,
				Versions:                map[string]*model.VersionMetadata{version: &versionData},
			}
		} else {
			projectResult[projectKey].Versions[version] = &versionData
		}
	}

	// set latest version
	for k, v := range result {
		var versions []*semver.Version
		for ver := range v.Versions {
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
	for k, v := range projectResult {
		var versions []*semver.Version
		for ver := range v.Versions {
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
			projectResult[k].Latest = latestVersion
		}
	}

	return result, projectResult, nil
}

func writeProjectIndexToFilesystem(outputDir string, data map[string]*model.ProjectMetadata) {
	var wmg sync.WaitGroup
	sem := make(chan struct{}, MaxConcurrency) // semaphore to limit concurrency
	for _, v := range data {
		current := v
		wmg.Add(1)
		sem <- struct{}{} // acquire semaphore

		go func(data *model.ProjectMetadata) {
			defer wmg.Done()
			defer func() {
				<-sem // release semaphore
			}()

			slog.Debug("writing artifact metadata", "group", data.GroupID, "artifact", data.ArtifactID)

			// write artifact data
			writeErr := util.WriteToFile(filepath.Join(outputDir, "project", strings.ReplaceAll(data.GroupID, ".", "/"), strings.ReplaceAll(data.ArtifactID, ".", "/"), "index.json"), data)
			if writeErr != nil {
				slog.Error("failed to write artifact metadata to file", "error", writeErr)
				os.Exit(1)
			}

			// write version data
			for version, versionMetadata := range data.Versions {
				writeVerErr := util.WriteToFile(filepath.Join(outputDir, "project", strings.ReplaceAll(data.GroupID, ".", "/"), strings.ReplaceAll(data.ArtifactID, ".", "/"), version+".json"), versionMetadata)
				if writeVerErr != nil {
					slog.Error("failed to write artifact metadata to file", "error", writeVerErr)
					os.Exit(1)
				}
			}

			// write badge data
			if data.Latest != "" {
				latestVersion, latestVersionFound := data.Versions[data.Latest]
				if latestVersionFound {
					badge := model.Badge{
						SchemaVersion: 1,
						Label:         "Reproducible Builds",
						LabelColor:    "1e5b96",
						Color:         util.Ternary(latestVersion.Reproducible, "ok", "error"),
						Message:       fmt.Sprintf("%s - %d/%d ok", data.Latest, latestVersion.ReproducibleFiles, latestVersion.ReproducibleFiles+latestVersion.NonReproducibleFiles),
						IsError:       !latestVersion.Reproducible,
						Style:         "flat",
					}
					writeErr = util.WriteToFile(filepath.Join(outputDir, "project", strings.ReplaceAll(data.GroupID, ".", "/"), strings.ReplaceAll(data.ArtifactID, ".", "/"), "badge.json"), badge)
					if writeErr != nil {
						slog.Error("failed to write artifact metadata to file", "error", writeErr)
						os.Exit(1)
					}
				}
			}
		}(current)
	}
	wmg.Wait()
}

func writeDependencyIndexToFilesystem(outputDir string, data map[string]*model.DependencyMetadata) {
	var wmg sync.WaitGroup
	sem := make(chan struct{}, MaxConcurrency) // semaphore to limit concurrency
	for _, v := range data {
		current := v
		wmg.Add(1)
		sem <- struct{}{} // acquire semaphore

		go func(data *model.DependencyMetadata) {
			defer wmg.Done()
			defer func() {
				<-sem // release semaphore
			}()

			slog.Debug("writing artifact metadata", "group", data.GroupID, "artifact", data.ArtifactID)

			// write artifact data
			writeErr := util.WriteToFile(filepath.Join(outputDir, "maven", strings.ReplaceAll(data.GroupID, ".", "/"), strings.ReplaceAll(data.ArtifactID, ".", "/"), "index.json"), data)
			if writeErr != nil {
				slog.Error("failed to write artifact metadata to file", "error", writeErr)
				os.Exit(1)
			}

			// write version data
			for version, versionMetadata := range data.Versions {
				writeVerErr := util.WriteToFile(filepath.Join(outputDir, "maven", strings.ReplaceAll(data.GroupID, ".", "/"), strings.ReplaceAll(data.ArtifactID, ".", "/"), version+".json"), versionMetadata)
				if writeVerErr != nil {
					slog.Error("failed to write artifact metadata to file", "error", writeVerErr)
					os.Exit(1)
				}
			}

			// write badge data
			if data.Latest != "" {
				latestVersion, latestVersionFound := data.Versions[data.Latest]
				if latestVersionFound {
					badge := model.Badge{
						SchemaVersion: 1,
						Label:         "Reproducible Builds",
						LabelColor:    "1e5b96",
						Color:         util.Ternary(latestVersion.Reproducible, "ok", "error"),
						Message:       fmt.Sprintf("%s - %d/%d ok", data.Latest, latestVersion.ReproducibleFiles, latestVersion.ReproducibleFiles+latestVersion.NonReproducibleFiles),
						IsError:       !latestVersion.Reproducible,
						Style:         "flat",
					}
					writeErr = util.WriteToFile(filepath.Join(outputDir, "maven", strings.ReplaceAll(data.GroupID, ".", "/"), strings.ReplaceAll(data.ArtifactID, ".", "/"), "badge.json"), badge)
					if writeErr != nil {
						slog.Error("failed to write artifact metadata to file", "error", writeErr)
						os.Exit(1)
					}
				}
			}
		}(current)
	}
	wmg.Wait()
}
