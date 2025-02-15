package schedulers

import "errors"

func list() map[string]string {
	return map[string]string{
          "scx_rusty": "ghcr.io/schedkit/scheds/scx_rusty:latest",
	}
}

func GetScheduler(id string) (string, error) {
	var image string

	for key, value := range list() {
		if key == id {
			image = value
		}
	}

	if len(image) == 0 {
		return "", errors.New("scheduler not found!")
	}

	return image, nil
}
