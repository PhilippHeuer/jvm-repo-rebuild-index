package util

import (
	"reflect"
	"testing"
)

func TestCoordinateToPath(t *testing.T) {
	tests := []struct {
		coordinate   string
		trimVersion  bool
		expectedPath string
	}{
		{
			coordinate:   "io.github.xanthic.cache:cache-bom:0.6.2",
			trimVersion:  true,
			expectedPath: "io/github/xanthic/cache/cache-bom",
		},
		{
			coordinate:   "io.github.xanthic.cache:cache-bom:0.6.2",
			trimVersion:  false,
			expectedPath: "io/github/xanthic/cache/cache-bom/0.6.2",
		},
		{
			coordinate:   "io.github.xanthic.cache:cache-bom",
			trimVersion:  true,
			expectedPath: "io/github/xanthic/cache/cache-bom",
		},
		{
			coordinate:   "io.github.xanthic.cache:cache-bom",
			trimVersion:  false,
			expectedPath: "io/github/xanthic/cache/cache-bom",
		},
	}

	for _, tt := range tests {
		t.Run(tt.coordinate, func(t *testing.T) {
			result := CoordinateToPath(tt.coordinate, tt.trimVersion)
			if result != tt.expectedPath {
				t.Errorf("CoordinateToPath(%q, %v) = %q, want %q", tt.coordinate, tt.trimVersion, result, tt.expectedPath)
			}
		})
	}
}

func TestParseMavenMetadata(t *testing.T) {
	xmlData := `
	<metadata>
	  <groupId>com.example.maven</groupId>
	  <artifactId>example-maven-plugin</artifactId>
	  <versioning>
		<latest>2.25</latest>
		<release>2.25</release>
		<versions>
		  <version>2.23</version>
		  <version>2.24</version>
		  <version>2.25</version>
		</versions>
		<lastUpdated>20241007071133</lastUpdated>
	  </versioning>
	</metadata>
	`

	expected := MavenMetadata{
		GroupID:    "com.example.maven",
		ArtifactID: "example-maven-plugin",
		Versioning: Versioning{
			Latest:  "2.25",
			Release: "2.25",
			Versions: []string{
				"2.23", "2.24", "2.25",
			},
			LastUpdated: "20241007071133",
		},
	}

	// Call the function
	result, err := ParseMavenMetadata([]byte(xmlData))
	if err != nil {
		t.Fatalf("ParseMavenMetadata returned an error: %v", err)
	}

	// Check if result matches expected
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("ParseMavenMetadata() = %v, want %v", result, expected)
	}
}
