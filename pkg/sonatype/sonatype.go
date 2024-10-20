package sonatype

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type DependencyRequest struct {
	Purl       string `json:"purl"`
	Page       int    `json:"page"`
	Size       int    `json:"size"`
	SearchTerm string `json:"searchTerm"`
}

type OssIndexInfo struct {
	URL                string `json:"url"`
	VulnerabilityCount *int   `json:"vulnerabilityCount"`
}

type BomDrInfo struct {
	URL string `json:"url"`
}

type Component struct {
	SourcePurl             string       `json:"sourcePurl"`
	DependencyPurl         string       `json:"dependencyPurl"`
	DependencyType         string       `json:"dependencyType"`
	DependencyRef          string       `json:"dependencyRef"`
	Scope                  string       `json:"scope"`
	SourceNamespace        string       `json:"sourceNamespace"`
	SourceName             string       `json:"sourceName"`
	SourceVersion          string       `json:"sourceVersion"`
	SourcePackaging        *string      `json:"sourcePackaging"`
	SourceOssIndexInfo     OssIndexInfo `json:"sourceOssIndexInfo"`
	SourceBomDrInfo        BomDrInfo    `json:"sourceBomDrInfo"`
	DependencyNamespace    string       `json:"dependencyNamespace"`
	DependencyName         string       `json:"dependencyName"`
	DependencyVersion      string       `json:"dependencyVersion"`
	DependencyPackaging    *string      `json:"dependencyPackaging"`
	DependencyClassifier   *string      `json:"dependencyClassifier"`
	DependencyOssIndexInfo OssIndexInfo `json:"dependencyOssIndexInfo"`
	DependencyBomDrInfo    BomDrInfo    `json:"dependencyBomDrInfo"`
	Description            string       `json:"description"`
	ChildCount             int          `json:"childCount"`
	Ingested               bool         `json:"ingested"`
	Licenses               []string     `json:"licenses"`
}

func ComponentEquals(a, b Component) bool {
	return a.DependencyNamespace == b.DependencyNamespace &&
		a.DependencyName == b.DependencyName &&
		a.DependencyVersion == b.DependencyVersion
}

type DependencyResponse struct {
	Components       []Component `json:"components"`
	Page             int         `json:"page"`
	PageSize         int         `json:"pageSize"`
	PageCount        int         `json:"pageCount"`
	TotalResultCount int         `json:"totalResultCount"`
	TotalCount       int         `json:"totalCount"`
}

func FetchDependencies(purl string, page, size int) (*DependencyResponse, error) {
	requestBody := DependencyRequest{
		Purl:       purl,
		Page:       page,
		Size:       size,
		SearchTerm: "",
	}
	body, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %v", err)
	}

	req, err := http.NewRequest("POST", "https://central.sonatype.com/api/internal/browse/dependencies", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	var result DependencyResponse
	err = json.Unmarshal(respBody, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %v", err)
	}

	return &result, nil
}

func FetchAllDependencies(purl string) ([]Component, error) {
	var allComponents []Component
	page := 0
	pageSize := 20

	for {
		response, err := FetchDependencies(purl, page, pageSize)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch dependencies on page %d: %v", page, err)
		}
		allComponents = append(allComponents, response.Components...)

		if len(response.Components) < pageSize {
			break
		}

		page++
	}

	return allComponents, nil
}
