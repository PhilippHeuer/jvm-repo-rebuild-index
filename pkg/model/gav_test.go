package model

import (
	"testing"
)

func TestGavPath(t *testing.T) {
	tests := []struct {
		coordinate   GAV
		trimVersion  bool
		expectedPath string
	}{
		{
			coordinate:   NewGAVIgnoreError("io.github.xanthic.cache:cache-bom:0.6.2"),
			trimVersion:  true,
			expectedPath: "io/github/xanthic/cache/cache-bom",
		},
		{
			coordinate:   NewGAVIgnoreError("io.github.xanthic.cache:cache-bom:0.6.2"),
			trimVersion:  false,
			expectedPath: "io/github/xanthic/cache/cache-bom/0.6.2",
		},
		{
			coordinate:   NewGAVIgnoreError("io.github.xanthic.cache:cache-bom"),
			trimVersion:  true,
			expectedPath: "io/github/xanthic/cache/cache-bom",
		},
		{
			coordinate:   NewGAVIgnoreError("io.github.xanthic.cache:cache-bom"),
			trimVersion:  false,
			expectedPath: "io/github/xanthic/cache/cache-bom",
		},
	}

	for _, tt := range tests {
		t.Run(tt.coordinate.Coordinate(), func(t *testing.T) {
			result := tt.coordinate.Path(tt.trimVersion)
			if result != tt.expectedPath {
				t.Errorf("GAV.Path(%q, %v) = %q, want %q", tt.coordinate, tt.trimVersion, result, tt.expectedPath)
			}
		})
	}
}
