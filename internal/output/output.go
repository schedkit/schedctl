package output

import "fmt"

func Out(output string, vars ...interface{}) (int, error) {
	return fmt.Printf(output, vars...) //nolint:forbidigo
}
