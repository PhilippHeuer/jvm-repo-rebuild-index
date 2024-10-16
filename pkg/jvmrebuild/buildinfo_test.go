package jvmrebuild

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParseBuildInfo(t *testing.T) {
	// Test cases
	tests := []struct {
		name     string
		filename string
		want     BuildInfo
		wantErr  bool
	}{
		{
			name:     "Maven - Single Output",
			filename: "testdata/maven-single.properties",
			want: BuildInfo{
				SpecVersion:  "1.0-SNAPSHOT",
				Name:         "jackson-databind",
				GroupID:      "com.fasterxml.jackson.core",
				ArtifactID:   "jackson-databind",
				Version:      "2.18.0",
				BuildTool:    "mvn",
				JavaVersion:  "8",
				OSName:       "Unix",
				SourceSCMUri: "scm:git:git@github.com:FasterXML/jackson-databind.git",
				SourceSCMTag: "jackson-databind-2.18.0",
				Outputs: []Output{
					{
						Coordinate: "com.fasterxml.jackson.core:jackson-databind",
						Files: map[string]File{
							"jackson-databind-2.18.0.pom": {
								Size:     "21547",
								Checksum: "abcdef",
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name:     "Gradle - Single Output",
			filename: "testdata/gradle-single.properties",
			want: BuildInfo{
				SpecVersion:  "",
				Name:         "com.github.philippheuer.credentialmanager:credentialmanager",
				GroupID:      "com.github.philippheuer.credentialmanager",
				ArtifactID:   "credentialmanager",
				BuildTool:    "gradle",
				JavaVersion:  "17",
				OSName:       "crlf-nogit",
				SourceSCMUri: "",
				SourceSCMTag: "",
				Outputs: []Output{
					{
						Coordinate: "com.github.philippheuer.credentialmanager:credentialmanager",
						Files: map[string]File{
							"credentialmanager-0.3.1-sources.jar": {
								Size:     "22338",
								Checksum: "17419eaa530f68941ab76a20c38c2a5e086bbf96cf18b1513edc0e50bee095c4f85b59818dd80607112c3e9fe76d20ff3b08ed55ec2eb4fa57c9003ace929220",
							},
							"credentialmanager-0.3.1.jar": {
								Size:     "40479",
								Checksum: "172521219799a119943d3704a676b637bd2b726c813c408593cad9e9f30bc11872d86987bd3311ae06d8ef5a557f1cc0aa4f7e58206f6df34a77e42504bf2a40",
							},
							"credentialmanager-0.3.1.pom": {
								Size:     "2881",
								Checksum: "44aa4192e9dc2a5f65fde2f56bdad6447208d1c9252be219728ef0e297a8782946a9395c8c2df8f6fe0986ed1d5bcddba1145456c31e2f24df3acae297cce0a6",
							},
						},
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, err := filepath.Abs(tt.filename)
			if err != nil {
				t.Fatalf("Failed to get absolute path: %v", err)
			}

			if _, err := os.Stat(path); os.IsNotExist(err) {
				t.Fatalf("Test file not found: %s", path)
			}

			got, err := ParseBuildInfo(path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseBuildInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("ParseBuildInfo() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
