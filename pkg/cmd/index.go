package cmd

import (
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"github.com/philippheuer/jvm-repo-rebuild-index/pkg/jvmrebuild"
	"github.com/philippheuer/jvm-repo-rebuild-index/pkg/model"
	"github.com/philippheuer/jvm-repo-rebuild-index/pkg/util"
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

func processFiles(files []string) (map[string]*model.Dependency, map[string]*model.Project) {
	depMetadata := make(map[string]*model.Dependency)
	projectMetadata := make(map[string]*model.Project)
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

func processFile(mvnMetadataFile string) (map[string]*model.Dependency, map[string]*model.Project, error) {
	result := make(map[string]*model.Dependency)     // individual artifact metadata
	projectResult := make(map[string]*model.Project) // project metadata

	dir := filepath.Dir(mvnMetadataFile)
	slog.Debug("found project", "path", mvnMetadataFile, "dir", filepath.Dir(mvnMetadataFile))

	// read maven-metadata.xml
	mvnMetadata, mvnMetadataErr := util.ParseMavenMetadataFile(mvnMetadataFile)
	if mvnMetadataErr != nil {
		return nil, nil, errors.Join(errors.New("failed to parse maven-metadata.xml"), mvnMetadataErr)
	}

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

		buildInfo, buildInfoErr := jvmrebuild.ParseBuildInfo(buildInfoFile)
		if buildInfoErr != nil {
			slog.Error("failed to parse buildinfo file", "error", buildInfoErr, "file", buildInfoFile)
			continue
		}
		if buildInfo.SpecVersion != "" && buildInfo.SpecVersion != "0.1-SNAPSHOT" && buildInfo.SpecVersion != "1.0-SNAPSHOT" {
			slog.Error("failed to find a supported artifactVersion in buildinfo file", "file", buildInfoFile, "version", buildInfo.SpecVersion)
			continue
		}

		buildCompareFile := strings.Replace(buildInfoFile, ".buildinfo", ".buildcompare", 1)
		buildCompare, buildCompareErr := util.ParseFile(buildCompareFile)
		if buildCompareErr != nil {
			slog.Error("failed to parse buildinfo file", "error", buildCompareErr, "file", buildCompareFile)
			continue
		}

		artifactVersion := buildCompare["version"] // buildInfo.Version is not always present, prefer buildCompare
		versionData := model.Version{
			Project:          buildInfo.Name,
			SCMUri:           buildInfo.SourceSCMUri,
			SCMTag:           buildInfo.SourceSCMTag,
			BuildTool:        buildInfo.BuildTool,
			BuildJavaVersion: buildInfo.JavaVersion,
			BuildOSName:      buildInfo.OSName,
			Reproducible:     buildCompare["ko"] == "0" && buildCompare["ok"] != "0",
			FileStats:        model.FileStats{},
		}
		allArtifacts := make(map[string]model.File)
		var allCoordinates []string

		var reproducibleFiles []string
		if rf, ok := buildCompare["okFiles"]; ok {
			reproducibleFiles = strings.Split(rf, " ")
		}
		slog.Debug("parsed buildinfo and buildcompare file", "file", buildInfoFile, "version", artifactVersion)

		// iterate over all outputs (look for key matching e.g. outputs.3.coordinates in buildInfo)
		for _, output := range buildInfo.Outputs {
			vd := versionData
			vd.Files = make(map[string]model.File)
			slog.Debug("found artifact", "coordinate", output.Coordinate)

			groupId := strings.SplitN(output.Coordinate, ":", 2)[0]
			artifactId := strings.SplitN(output.Coordinate, ":", 2)[1]
			if groupId == "" || artifactId == "" {
				slog.Warn("no group-id or artifact-id found for version", "version", artifactVersion, "coordinate", output.Coordinate)
				continue
			}

			for name, file := range output.Files {
				if strings.HasPrefix(name, artifactId+"-"+artifactVersion) {
					reproducible := slices.Contains(reproducibleFiles, name)
					vd.Files[name] = model.File{
						Size:         file.Size,
						Checksum:     file.Checksum,
						Reproducible: reproducible,
					}
					allArtifacts[name] = vd.Files[name]
				}
			}
			vd.SetModuleFileStats()

			// create or append to result
			if _, ok := result[output.Coordinate]; !ok {
				result[output.Coordinate] = &model.Dependency{
					RebuildProjectUrl: overviewUrl,
					GroupID:           groupId,
					ArtifactID:        artifactId,
					Versions:          map[string]*model.Version{artifactVersion: &vd},
					Latest:            mvnMetadata.Versioning.Latest,
				}
				allCoordinates = append(allCoordinates, output.Coordinate)
			} else {
				result[output.Coordinate].Versions[artifactVersion] = &vd
			}
		}

		// set reproducible file count for the entire project
		for rk := range result {
			if result[rk].Versions[artifactVersion] == nil {
				continue
			}

			result[rk].Versions[artifactVersion].SetTotalFileStats(allArtifacts)
		}

		// append project metadata
		versionData.Files = allArtifacts
		versionData.SetTotalFileStats(allArtifacts)
		versionData.SetModuleFileStats()
		projectKey := buildInfo.GroupID + ":" + buildInfo.ArtifactID
		if _, ok := projectResult[projectKey]; !ok {
			projectResult[projectKey] = &model.Project{
				RebuildProjectUrl: overviewUrl,
				GroupID:           buildInfo.GroupID,
				ArtifactID:        buildInfo.ArtifactID,
				Modules:           allCoordinates,
				Versions:          map[string]*model.Version{artifactVersion: &versionData},
				Latest:            mvnMetadata.Versioning.Latest,
			}
		} else {
			projectResult[projectKey].Versions[artifactVersion] = &versionData
		}
	}

	return result, projectResult, nil
}

func writeProjectIndexToFilesystem(outputDir string, data map[string]*model.Project) {
	var wmg sync.WaitGroup
	sem := make(chan struct{}, MaxConcurrency) // semaphore to limit concurrency
	for _, v := range data {
		current := v
		wmg.Add(1)
		sem <- struct{}{} // acquire semaphore

		go func(data *model.Project) {
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
			/* TODO: enable or remove, i prefer rendering badges via api
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
			*/
		}(current)
	}
	wmg.Wait()
}

func writeDependencyIndexToFilesystem(outputDir string, data map[string]*model.Dependency) {
	var wmg sync.WaitGroup
	sem := make(chan struct{}, MaxConcurrency) // semaphore to limit concurrency
	for _, v := range data {
		current := v
		wmg.Add(1)
		sem <- struct{}{} // acquire semaphore

		go func(data *model.Dependency) {
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
			/* TODO: enable or remove, i prefer rendering badges via api
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
			*/
		}(current)
	}
	wmg.Wait()
}
