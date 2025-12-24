package schedulers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"schedctl/internal/output"
)

const remoteURL = "https://raw.githubusercontent.com/schedkit/plumbing/refs/heads/main/manifest.json"

type ManifestEntry struct {
	ImageURI string `json:"imageURI"`
}

type Manifest map[string]ManifestEntry

type ImageSource string

const (
	SourceManifest ImageSource = "manifest"
	SourceDirect   ImageSource = "direct"
)

type SchedulerImage struct {
	ImageURI string
	Source   ImageSource
}

func List() Manifest {
	var manifest Manifest

	resp, err := http.Get(remoteURL)
	if err != nil {
		output.Error("Failed to fetch remote file: %v", err)
		return manifest
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		output.Error("Failed to read response body: %v", err)
		return manifest
	}

	if err := json.Unmarshal(body, &manifest); err != nil {
		output.Error("Failed to unmarshal JSON: %v", err)
		return manifest
	}

	return manifest
}

// isContainerImage checks if a string looks like a container image URI.
// Examples:
//   - ghcr.io/user/repo
//   - docker.io/nginx:latest
//   - quay.io/org/image:v1.0
//   - localhost:5000/myimage
//   - nginx/alpine
func isContainerImage(input string) bool {
	if strings.Contains(input, "/") {
		parts := strings.Split(input, "/")
		firstPart := parts[0]

		if strings.Contains(firstPart, ".") || strings.Contains(firstPart, ":") {
			return true
		}

		if len(parts) >= 2 {
			return true
		}
	}

	return false
}

func ensureImageTag(image string) string {
	if strings.Contains(image, "@") {
		return image
	}

	parts := strings.Split(image, "/")
	lastPart := parts[len(parts)-1]

	// Check if the last part (repo name) has a tag
	if !strings.Contains(lastPart, ":") {
		return image + ":latest"
	}

	return image
}

func GetScheduler(id string) (SchedulerImage, error) {
	var result SchedulerImage

	manifest := List()
	for key, entry := range manifest {
		if key == id {
			result.ImageURI = ensureImageTag(entry.ImageURI)
			result.Source = SourceManifest
			return result, nil
		}
	}

	if isContainerImage(id) {
		result.ImageURI = ensureImageTag(id)
		result.Source = SourceDirect
		return result, nil
	}

	return result, errors.New(
		"scheduler not found in manifest and input does not appear to be a valid container image URI",
	)
}
