package schedulers

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/go-containerregistry/pkg/crane"
)

type ImageInfo struct {
	ImageRef     string
	Digest       string
	Created      *time.Time
	Architecture string
	OS           string
	Labels       map[string]string
}

type manifestMetadata struct {
	Annotations map[string]string `json:"annotations,omitempty"`
}

type configMetadata struct {
	Created      *time.Time `json:"created,omitempty"`
	Architecture string     `json:"architecture,omitempty"`
	OS           string     `json:"os,omitempty"`
	Config       struct {
		Labels map[string]string `json:"Labels,omitempty"`
	} `json:"config,omitempty"`
}

func Inspect(imageRef string) (ImageInfo, error) {
	info := ImageInfo{ImageRef: imageRef, Labels: map[string]string{}}

	digest, err := crane.Digest(imageRef)
	if err != nil {
		return info, fmt.Errorf("resolve digest for %q: %w", imageRef, err)
	}
	info.Digest = digest

	configRaw, err := crane.Config(imageRef)
	if err != nil {
		return info, fmt.Errorf("fetch image config for %q: %w", imageRef, err)
	}
	var cfg configMetadata
	if err := json.Unmarshal(configRaw, &cfg); err != nil {
		return info, fmt.Errorf("parse image config for %q: %w", imageRef, err)
	}
	info.Created = cfg.Created
	info.Architecture = cfg.Architecture
	info.OS = cfg.OS
	for k, v := range cfg.Config.Labels {
		info.Labels[k] = v
	}

	manifestRaw, err := crane.Manifest(imageRef)
	if err != nil {
		return info, fmt.Errorf("fetch image manifest for %q: %w", imageRef, err)
	}
	var mf manifestMetadata
	if err := json.Unmarshal(manifestRaw, &mf); err != nil {
		return info, fmt.Errorf("parse image manifest for %q: %w", imageRef, err)
	}
	for k, v := range mf.Annotations {
		info.Labels[k] = v
	}

	return info, nil
}
