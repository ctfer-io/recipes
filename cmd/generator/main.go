package main

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi/sdk/v3/go/common/workspace"
	"go.uber.org/multierr"
	"gopkg.in/yaml.v3"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/file"
	"oras.land/oras-go/v2/content/oci"
)

const (
	// GOPROXY=${goproxy} helps in making the binary reproducible
	goproxy = "https://proxy.golang.org,direct"

	fileType     = "application/vnd.ctfer-io.file"
	scenarioType = "application/vnd.ctfer-io.scenario"
	dist         = "dist"
)

var (
	ecosystems = []string{
		"chall-manager",
	}
)

func main() {
	if err := run(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] %s\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	// Create root directory in which to export OCI recipes
	_ = os.Mkdir(dist, os.ModePerm)

	ver := os.Getenv("VERSION")

	// Find entries to build
	for _, eco := range ecosystems {
		entries, err := os.ReadDir(eco)
		if err != nil {
			return err
		}
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}

			dir := filepath.Join(eco, e.Name())
			fmt.Printf("[+] Building %s\n", dir)

			into := filepath.Join(dist, fmt.Sprintf("%s_%s_%s.oci.tar.gz", eco, e.Name(), ver))
			if err := build(ctx, dir, into, ver); err != nil {
				return errors.Wrapf(err, "failed to build %s", dir)
			}
			fmt.Printf("    Exported to %s\n", into)
		}
	}

	return nil
}

type BuildEntry struct {
	Path   string
	Digest string
}

func build(ctx context.Context, dir, into, ver string) error {
	// Compile Go binary
	if err := compile(ctx, dir); err != nil {
		return err
	}

	// Then pack it all in an OCI layout in filesystem
	if err := ociLayout(ctx, dir, ver); err != nil {
		return err
	}

	// Compress it in a tar.gz and compute its sha256 sum
	return compress(dir, into)
}

func compile(ctx context.Context, dir string) error {
	cmd := exec.CommandContext(ctx, "go", "build", "-o", "main", "main.go")
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("GOPROXY=%s", goproxy),
	)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return errors.Wrapf(err, "output: %s", out)
	}
	return nil
}

func ociLayout(ctx context.Context, dir, ver string) error {
	// Create new file store
	store, err := file.New(dir)
	if err != nil {
		return errors.Wrapf(err, "creating file store in %s", dir)
	}
	defer func() { _ = store.Close() }()

	// Add files
	if err := preparePulumiYaml(dir); err != nil {
		return errors.Wrap(err, "preparing Pulumi.yaml")
	}
	layers := []ocispec.Descriptor{}
	for _, f := range []string{"main", "Pulumi.yaml"} {
		desc, err := store.Add(ctx, f, fileType, f)
		if err != nil {
			return errors.Wrapf(err, "adding file %s to ORAS file store", f)
		}
		layers = append(layers, desc)
	}

	// Pack the manifest in store
	root, err := oras.PackManifest(ctx, store, oras.PackManifestVersion1_1, scenarioType, oras.PackManifestOptions{
		Layers: layers,
	})
	if err != nil {
		return errors.Wrap(err, "packing manifest")
	}

	// Tag the memory store
	if err := store.Tag(ctx, root, ver); err != nil {
		return errors.Wrap(err, "tagging memory store")
	}

	// Create a new OCI layout in filesystem
	odir := filepath.Join(dir, dist)
	dst, err := oci.New(odir)
	if err != nil {
		return errors.Wrapf(err, "creating new OCI registry in %s", odir)
	}

	// Copy content (graph)
	if err := oras.CopyGraph(ctx, store, dst, root, oras.DefaultCopyOptions.CopyGraphOptions); err != nil {
		return errors.Wrapf(err, "copying graph into %s", odir)
	}

	return nil
}

func preparePulumiYaml(dir string) error {
	pyp := filepath.Join(dir, "Pulumi.yaml")
	b, err := os.ReadFile(pyp)
	if err != nil {
		return err
	}

	var proj workspace.Project
	if err := yaml.Unmarshal(b, &proj); err != nil {
		return errors.Wrap(err, "unmarshalling Pulumi.yaml")
	}
	if _, ok := proj.Runtime.Options()["binary"]; !ok {
		proj.Runtime.SetOption("binary", "./main")
	}

	f, err := os.OpenFile(pyp, os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	enc := yaml.NewEncoder(f)
	enc.SetIndent(2) // common practice through ctfer-io codebases
	if err := enc.Encode(proj); err != nil {
		return errors.Wrap(multierr.Append(
			err,
			f.Close(),
		), "marshalling Pulumi.yaml")
	}
	return f.Close()
}

func compress(path, target string) error {
	tarfile, err := os.Create(target)
	if err != nil {
		return errors.Wrapf(err, "creating tar.gz %s", target)
	}

	// Create cascading writers
	gzipWriter := gzip.NewWriter(tarfile)
	tarWriter := tar.NewWriter(gzipWriter)

	dir := filepath.Join(path, dist)
	err = filepath.Walk(dir, func(file string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Compute the relative path from the source directory
		relPath, err := filepath.Rel(dir, file)
		if err != nil {
			return err
		}

		// Ensure we skip the root directory
		if relPath == "." {
			return nil
		}

		// Open file if it's not a directory
		var fileReader io.Reader
		var fileHeader *tar.Header

		if fi.Mode().IsRegular() {
			f, err := os.Open(file)
			if err != nil {
				return err
			}
			defer f.Close()
			fileReader = f

			fileHeader, err = tar.FileInfoHeader(fi, "")
			if err != nil {
				return err
			}
			fileHeader.Name = relPath
		} else if fi.IsDir() {
			fileHeader = &tar.Header{
				Name:     relPath + "/",
				Typeflag: tar.TypeDir,
				Mode:     int64(fi.Mode()),
				ModTime:  fi.ModTime(),
			}
		} else {
			// Ignore symlinks, devices, etc.
			return nil
		}

		if err := tarWriter.WriteHeader(fileHeader); err != nil {
			return fmt.Errorf("failed to write tar header: %w", err)
		}

		if fileReader != nil {
			if _, err := io.Copy(tarWriter, fileReader); err != nil {
				return fmt.Errorf("failed to write file data: %w", err)
			}
		}

		return nil
	})
	if err != nil {
		return errors.Wrapf(err, "creating tar.gz %s", target)
	}

	// Close all writers and file
	if err := tarWriter.Close(); err != nil {
		return err
	}
	if err := gzipWriter.Close(); err != nil {
		return err
	}
	return tarfile.Close()
}
