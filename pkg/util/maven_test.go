package util

import (
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
