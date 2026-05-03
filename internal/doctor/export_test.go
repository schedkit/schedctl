package doctor

// Forward internal helpers for tests in the external doctor_test package.

func ParseKernelVersion(s string) (int, int, error) { return parseKernelVersion(s) }
func SocketReachable(path string) bool              { return socketReachable(path) }
