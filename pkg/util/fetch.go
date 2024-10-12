package util

import (
	"encoding/json"
	"net/http"
	"os"
)

func LoadFromURL[T any](url string) (T, error) {
	var result T

	resp, err := http.Get(url)
	if err != nil {
		return result, err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return result, err
	}

	return result, nil
}

func LoadFromDisk[T any](filename string) (T, error) {
	var result T

	content, err := os.ReadFile(filename)
	if err != nil {
		return result, err
	}

	err = json.Unmarshal(content, &result)
	if err != nil {
		return result, err
	}

	return result, nil
}
