package util

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/charlievieth/fastwalk"
)

func FindFiles(rootPath string, suffix string) ([]string, error) {
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

func ParseFile(filename string) (map[string]string, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return ParseProperties(string(content)), nil
}

func ParseProperties(content string) map[string]string {
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

func WriteToFile(filename string, data any) error {
	if err := os.MkdirAll(filepath.Dir(filename), os.ModePerm); err != nil {
		return err
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	// encoder.SetIndent("", "  ") // pretty print
	return encoder.Encode(data)
}

func Ternary[T any](condition bool, trueVal T, falseVal T) T {
	if condition {
		return trueVal
	}
	return falseVal
}

func TrimURLProtocolAndTrailingSlash(url string) string {
	url = strings.TrimPrefix(url, "http://")
	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimRight(url, "/")
	return url
}
