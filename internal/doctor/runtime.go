package doctor

import (
	"net"
	"strings"
	"time"
)

func runtimeChecks() []Check {
	podmanPaths := []string{"/run/podman/podman.sock"}
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
					if socketReachable(p) {
						return StatusPass, p
					}
				}
				return StatusFail, "neither Podman nor containerd socket is reachable"
			},
		},
	}
}

func socketCheckFunc(candidates []string) CheckFunc {
	return func() (Status, string) {
		for _, p := range candidates {
			if socketReachable(p) {
				return StatusPass, p
			}
		}
		return StatusFail, "none reachable: " + strings.Join(candidates, ", ")
	}
}

func socketReachable(path string) bool {
	conn, err := net.DialTimeout("unix", path, time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}
