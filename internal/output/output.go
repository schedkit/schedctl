package output

import (
	"fmt"
	"io"
)

func Out(output string, vars ...interface{}) (int, error) {
	return fmt.Printf(output, vars...) //nolint:forbidigo
}

func Outf(w io.Writer, format string, vars ...interface{}) (int, error) {
	return fmt.Fprintf(w, format, vars...)
}

func Outln(w io.Writer, vars ...interface{}) (int, error) {
	return fmt.Fprintln(w, vars...)
}

func Error(output string, vars ...interface{}) {
	_ = fmt.Errorf(output, vars...)
}
