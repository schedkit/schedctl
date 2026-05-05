package schedulers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/google/go-containerregistry/pkg/crane"

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

func imageRepo(image string) string {
	if idx := strings.Index(image, "@"); idx != -1 {
		return image[:idx]
	}

	parts := strings.Split(image, "/")
	last := parts[len(parts)-1]

	if idx := strings.Index(last, ":"); idx != -1 {
		parts[len(parts)-1] = last[:idx]
	}

	return strings.Join(parts, "/")
}

func GetScheduler(id, version string) (SchedulerImage, error) {
	var result SchedulerImage

	manifest := List()
	for key, entry := range manifest {
		if key == id {
			result.ImageURI = resolveImageVersion(entry.ImageURI, version)
			result.Source = SourceManifest
			return result, nil
		}
	}

	if isContainerImage(id) {
		result.ImageURI = resolveImageVersion(id, version)
		result.Source = SourceDirect
		return result, nil
	}

	return result, errors.New(
		"scheduler not found in manifest and input does not appear to be a valid container image URI",
	)
}

func resolveImageVersion(image, version string) string {
	if version != "" {
		return imageRepo(image) + ":" + version
	}

	return ensureImageTag(image)
}

func ListVersions(id string) ([]string, error) {
	manifest := List()
	for key, entry := range manifest {
		if key == id {
			return crane.ListTags(imageRepo(entry.ImageURI))
		}
	}

	if isContainerImage(id) {
		return crane.ListTags(imageRepo(id))
	}

	return nil, errors.New(
		"scheduler not found in manifest and input does not appear to be a valid container image URI",
	)
}
