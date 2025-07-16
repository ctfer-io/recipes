package main

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"
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
	if err := os.Mkdir(dist, os.ModePerm); err != nil {
		return err
	}

	ver := os.Getenv("VERSION")

	// Find entries to build
	bes := []*BuildEntry{}
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

			into := filepath.Join(dist, fmt.Sprintf("recipes_%s_%s_%s.oci.tar.gz", eco, e.Name(), ver))
			dig, err := build(ctx, dir, into)
			if err != nil {
				return errors.Wrapf(err, "failed to build %s", dir)
			}
			bes = append(bes, &BuildEntry{
				Path:   into,
				Digest: dig,
			})
		}
	}

	// Advertise them through an output, later passed to SLSA generator
	return advertise(bes)
}

type BuildEntry struct {
	Path   string
	Digest string
}

func build(ctx context.Context, dir, into string) (string, error) {
	// Compile Go binary
	if err := compile(ctx, dir); err != nil {
		return "", err
	}

	// Then pack it all in an OCI layout in filesystem
	if err := ociLayout(ctx, dir); err != nil {
		return "", err
	}

	// Compress it in a tar.gz and compute its sha256 sum
	digest, err := compress(dir, into)
	if err != nil {
		return "", err
	}
	return digest, nil
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

func ociLayout(ctx context.Context, dir string) error {
	// Create new file store
	store, err := file.New(dir)
	if err != nil {
		return errors.Wrapf(err, "creating file store in %s", dir)
	}
	defer func() { _ = store.Close() }()

	// Add files
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
	if err := store.Tag(ctx, root, root.Digest.String()); err != nil {
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

func compress(path, target string) (string, error) {
	tarfile, err := os.Create(target)
	if err != nil {
		return "", errors.Wrapf(err, "creating tar.gz %s", target)
	}
	defer tarfile.Close()

	hasher := sha256.New()

	// Create cascading writers
	multiWriter := io.MultiWriter(tarfile, hasher)
	gzipWriter := gzip.NewWriter(multiWriter)
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
		return "", errors.Wrapf(err, "creating tar.gz %s", target)
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func advertise(bes []*BuildEntry) error {
	var entries []string
	for _, be := range bes {
		entry := fmt.Sprintf("path=%s,digest=sha256:%s\n", be.Path, be.Digest)
		b64 := base64.StdEncoding.EncodeToString([]byte(entry))
		entries = append(entries, b64)
	}
	return output("hashes", strings.Join(entries, "\n"))
}

func output(k, v string) error {
	// Open GitHub output file
	f, err := os.OpenFile(os.Getenv("GITHUB_OUTPUT"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0777)
	if err != nil {
		return errors.Wrap(err, "opening action output file")
	}
	defer f.Close()

	// Write and ensure it went fine
	if _, err = fmt.Fprintf(f, "%s=%s\n", k, v); err != nil {
		return errors.Wrapf(err, "writing %s output", k)
	}
	return nil
}
