package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	docs "github.com/urfave/cli-docs/v3"
	"github.com/urfave/cli/v3"

	cmd "schedctl/cmd/schedctl"
)

func main() {
	out := flag.String("out", "dist/man", "output directory")
	flag.Parse()

	if err := os.MkdirAll(*out, 0o755); err != nil {
		fail(err)
	}

	extrasDir := extrasDir()

	root := cmd.NewRootCmd()

	if err := writeMan(root, "schedctl", filepath.Join(*out, "schedctl.1"), extrasDir); err != nil {
		fail(err)
	}

	persistent := persistentFlags(root)

	for _, sub := range root.Commands {
		name := "schedctl-" + sub.Name
		sub.Name = name
		sub.UsageText = "schedctl " + strings.TrimPrefix(name, "schedctl-")
		if sub.ArgsUsage != "" {
			sub.UsageText += " " + sub.ArgsUsage
		}
		sub.Flags = append(sub.Flags, persistent...)
		if err := writeMan(sub, name, filepath.Join(*out, name+".1"), extrasDir); err != nil {
			fail(err)
		}
	}
}

func writeMan(c *cli.Command, page, path, extrasDir string) error {
	body, err := docs.ToManWithSection(c, 1)
	if err != nil {
		return fmt.Errorf("render %s: %w", page, err)
	}
	extras, err := readExtras(extrasDir, page)
	if err != nil {
		return err
	}
	if extras != "" {
		if !strings.HasSuffix(body, "\n") {
			body += "\n"
		}
		body += extras
	}
	return os.WriteFile(path, []byte(body), 0o644) //nolint:gosec
}

func readExtras(dir, page string) (string, error) {
	path := filepath.Join(dir, page+".troff")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("read extras %s: %w", path, err)
	}
	return string(data), nil
}

func persistentFlags(c *cli.Command) []cli.Flag {
	var out []cli.Flag
	for _, f := range c.Flags {
		if lf, ok := f.(interface{ IsLocal() bool }); ok && lf.IsLocal() {
			continue
		}
		out = append(out, f)
	}
	return out
}

func extrasDir() string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return "extras"
	}
	return filepath.Join(filepath.Dir(file), "extras")
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "gen-man:", err)
	os.Exit(1)
}
