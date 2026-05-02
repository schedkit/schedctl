package doctor

import (
	"os"
	"strconv"
	"strings"
)

func runtimeChecks() []Check {
	podmanPaths := podmanSocketCandidates()
	containerdPaths := []string{"/run/containerd/containerd.sock"}

	return []Check{
		{
			ID:          "runtime.podman_socket",
			Description: "Podman API socket is reachable",
			Severity:    SeverityWarn,
			Remediation: "start podman.socket (e.g. `systemctl --user enable --now podman.socket`) or rely on containerd",
			Func:        socketCheckFunc(podmanPaths),
		},
		{
			ID:          "runtime.containerd_socket",
			Description: "containerd API socket is reachable",
			Severity:    SeverityWarn,
			Remediation: "start containerd (e.g. `systemctl enable --now containerd`) or rely on Podman",
			Func:        socketCheckFunc(containerdPaths),
		},
		{
			ID:          "runtime.any",
			Description: "at least one supported container runtime is reachable",
			Severity:    SeverityError,
			Remediation: "start either Podman or containerd before running schedulers",
			Func: func() (Status, string) {
				all := append([]string{}, podmanPaths...)
				all = append(all, containerdPaths...)
				for _, p := range all {
					if isSocket(p) {
						return StatusPass, p
					}
				}
				return StatusFail, "neither Podman nor containerd socket is reachable"
			},
		},
	}
}

func podmanSocketCandidates() []string {
	paths := []string{"/run/podman/podman.sock"}
	if rt := os.Getenv("XDG_RUNTIME_DIR"); rt != "" {
		paths = append(paths, rt+"/podman/podman.sock")
	} else {
		paths = append(paths, "/run/user/"+strconv.Itoa(os.Getuid())+"/podman/podman.sock")
	}
	return paths
}

func socketCheckFunc(candidates []string) CheckFunc {
	return func() (Status, string) {
		for _, p := range candidates {
			if isSocket(p) {
				return StatusPass, p
			}
		}
		return StatusFail, "none reachable: " + strings.Join(candidates, ", ")
	}
}

func isSocket(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeSocket != 0
}
