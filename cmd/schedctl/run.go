package cmd

import (
	"fmt"
	"github.com/opencontainers/umoci"
	"github.com/opencontainers/umoci/oci/cas/dir"
	"github.com/opencontainers/umoci/oci/casext"
	"github.com/opencontainers/umoci/oci/layer"

	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"

	"github.com/spf13/cobra"
)

func NewRunCmd() *cobra.Command {
	startCmd := &cobra.Command{
		Use:   "run",
		Short: "Run a specific scheduler",
		RunE:  run,
	}

	return startCmd
}

func run(cmd *cobra.Command, arguments []string) error {
	src := cmd.Flags().Args()[0]
	bundlePath := "kekw2"

	opt := crane.GetOptions()

	ref, err := name.ParseReference(src, opt.Name...)
	if err != nil {
		return fmt.Errorf("parsing reference %q: %w", src, err)
	}

	img, err := remote.Image(ref, opt.Remote...)

	if err != nil {
		return fmt.Errorf("error: %w", err)
	}

	size, _ := img.Size()

	fmt.Println("%v", size)

	crane.SaveOCI(img, "kekw")

	var unpackOptions layer.UnpackOptions
	var meta umoci.Meta
	meta.Version = umoci.MetaVersion
	meta.Version = umoci.MetaVersion

	engine, err := dir.Open("kekw")
	if err != nil {
		return fmt.Errorf("open CAS: %w", err)
	}
	engineExt := casext.NewEngine(engine)
	defer engine.Close()
        err = umoci.Unpack(engineExt, "latest", bundlePath, unpackOptions)
	if err != nil {
		fmt.Println("Error unpacking image:", err)
		return nil
	}

	return nil
}

func BoolPointer(b bool) *bool {
	return &b
}
