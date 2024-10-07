package util

import (
	"strings"
	"unicode"
)

func IsValidMavenCoordinate(coordinate string) bool {
	for _, ch := range coordinate {
		if !unicode.IsLetter(ch) && !unicode.IsDigit(ch) && ch != '.' && ch != ':' && ch != '-' {
			return false
		}
	}
	return true
}

func CoordinateToPath(coordinate string, trimVersion bool) string {
	parts := strings.SplitN(coordinate, ":", 3)
	groupAndArtifact := strings.NewReplacer(".", "/", ":", "/").Replace(parts[0] + ":" + parts[1])

	if trimVersion || len(parts) < 3 {
		return groupAndArtifact
	}
	return groupAndArtifact + "/" + parts[2]
}
