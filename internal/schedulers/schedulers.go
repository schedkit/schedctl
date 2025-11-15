package schedulers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"schedctl/internal/output"
)

const remoteURL = "https://raw.githubusercontent.com/schedkit/plumbing/refs/heads/main/manifest.json"

type ManifestEntry struct {
	ImageURI string `json:"imageURI"`
}

type Manifest map[string]ManifestEntry

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

func GetScheduler(id string) (string, error) {
	var image string

	for key, entry := range List() {
		if key == id {
			// For the moment we just append the :latest tag to the image
			image = entry.ImageURI + ":latest"
		}
	}

	if len(image) == 0 {
		return "", errors.New("scheduler not found")
	}

	return image, nil
}
