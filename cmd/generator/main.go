package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi/sdk/v3/go/common/workspace"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.uber.org/multierr"
	"gopkg.in/yaml.v3"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/file"
	"oras.land/oras-go/v2/content/oci"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
	"oras.land/oras-go/v2/registry/remote/retry"
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

	preparedFiles = []string{
		"main",
		"Pulumi.yaml",
	}

	dhClient *DockerHubClient
	dhPat    string
)

func main() {
	if err := run(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] %s\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) (err error) {
	// Login to Docker Hub
	dhPat = strings.TrimSpace(os.Getenv("DOCKERHUB_PAT"))
	if dhPat == "" {
		log.Fatal("Docker Hub PAT token is empty...")
	}

	dhClient, err = Login(ctx, "ctferio", dhPat)
	if err != nil {
		return err
	}

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
			// Skip commonly-used (shared) datastructures and helpers
			if e.Name() == "common" {
				continue
			}

			dir := filepath.Join(eco, e.Name())
			fmt.Printf("[+] Building %s@%s\n", dir, ver)

			// Into which compressed archive
			into := filepath.Join(dist, fmt.Sprintf("%s_%s_%s.tar.gz", eco, e.Name(), ver))

			// Transform into a Docker-compliant name
			sub := e.Name()
			sub = strings.ToLower(sub)
			sub = strings.NewReplacer(
				".", "-",
			).Replace(sub)
			dhRepoName := fmt.Sprintf("recipes_%s_%s", eco, sub)

			if err := build(ctx, dir, into, dhRepoName, ver); err != nil {
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

func build(ctx context.Context, dir, into, dhRepoName, ver string) error {
	// Compile Go binary
	if err := compile(ctx, dir); err != nil {
		return err
	}

	// Then pack it all in an OCI layout in filesystem ...
	if err := ociLayout(ctx, dir, ver); err != nil {
		return err
	}
	// ... and push it to Docker Hub
	if err := dhubPush(ctx, dir, dhRepoName, ver); err != nil {
		return err
	}

	// ... and compress it in a tag.gz
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
	// Prepare the Pulumi.yaml file with the prebuilt content
	if err := preparePulumiYaml(dir); err != nil {
		return errors.Wrap(err, "preparing Pulumi.yaml")
	}

	// Copy prepared data into a clean directory
	tmpDir := filepath.Join(os.TempDir(), dir)
	if err := os.MkdirAll(tmpDir, os.ModePerm); err != nil {
		return err
	}

	for _, f := range preparedFiles {
		if err := copyInto(filepath.Join(dir, f), tmpDir); err != nil {
			return err
		}
	}

	// Create new file fs
	fs, err := file.New(tmpDir)
	if err != nil {
		return errors.Wrapf(err, "creating file store in %s", dir)
	}
	defer func() { _ = fs.Close() }()

	// Add files
	layers := []ocispec.Descriptor{}
	for _, f := range preparedFiles {
		layer, err := fs.Add(ctx, f, fileType, "")
		if err != nil {
			return errors.Wrapf(err, "adding file %s to ORAS file store", f)
		}
		layers = append(layers, layer)
	}

	// Pack the manifest in store
	root, err := oras.PackManifest(ctx, fs,
		oras.PackManifestVersion1_1,
		scenarioType,
		oras.PackManifestOptions{Layers: layers})
	if err != nil {
		return errors.Wrap(err, "packing manifest")
	}

	// Tag the memory store
	fmt.Printf("    Digest: %s\n", root.Digest)
	if err := fs.Tag(ctx, root, root.Digest.String()); err != nil {
		return errors.Wrap(err, "tagging memory store")
	}

	// Create a new OCI layout in filesystem
	odir := filepath.Join(dir, dist)
	dst, err := oci.New(odir)
	if err != nil {
		return errors.Wrapf(err, "creating new OCI registry in %s", odir)
	}

	// Copy content (graph)
	if _, err := oras.Copy(ctx, fs, root.Digest.String(), dst, ver, oras.DefaultCopyOptions); err != nil {
		return errors.Wrapf(err, "copying into %s", odir)
	}

	return nil
}

func copyInto(path, dir string) error {
	src, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() {
		_ = src.Close()
	}()

	dst, err := os.Create(filepath.Join(dir, filepath.Base(path)))
	if err != nil {
		return err
	}
	defer func() {
		_ = dst.Close()
	}()

	_, err = io.Copy(dst, src)
	return err
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

	// Compress the prepared files
	for _, pf := range preparedFiles {
		fpath := filepath.Join(path, pf)

		f, err := os.Open(fpath)
		if err != nil {
			return err
		}
		defer func() {
			_ = f.Close()
		}()

		fi, err := os.Stat(fpath)
		if err != nil {
			return err
		}
		fileHeader, err := tar.FileInfoHeader(fi, "")
		if err != nil {
			return err
		}
		if err := tarWriter.WriteHeader(fileHeader); err != nil {
			return errors.Wrapf(err, "failed to write tar header of %s", pf)
		}

		if _, err := io.Copy(tarWriter, f); err != nil {
			return errors.Wrapf(err, "failed to copy %s", pf)
		}
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

func dhubPush(ctx context.Context, dir, repoName, version string) error {
	// Create the repository if does not exist already
	if err := dhClient.UpsertRepo(ctx, dir, repoName); err != nil {
		return errors.Wrapf(err, "upserting ctferio/%s", repoName)
	}

	// Load OCI layout that previous steps built
	odir := filepath.Join(dir, dist)
	ociLayout, err := oci.New(odir)
	if err != nil {
		return err
	}

	// Will be uploaded with this reference
	ref := fmt.Sprintf("docker.io/ctferio/%s:%s", repoName, version)

	// Then copy to remote
	repo, err := remote.NewRepository(ref)
	if err != nil {
		return err
	}
	repo.Client = &auth.Client{
		Cache: auth.NewCache(),
		Client: &http.Client{
			Transport: otelhttp.NewTransport(retry.NewTransport(nil)),
		},
		Credential: auth.StaticCredential("docker.io", auth.Credential{
			Username: "ctferio",
			Password: dhPat,
		}),
	}

	fmt.Printf("    Pushing %s\n", ref)
	if _, err := oras.Copy(ctx,
		ociLayout, version, // from OCI layout
		repo, version, // to DockerHub
		oras.DefaultCopyOptions,
	); err != nil {
		return err
	}

	// And delete the oci directory so it is cleaned up for compression
	return os.RemoveAll(odir)
}

type DockerHubClient struct {
	token string
}

func Login(ctx context.Context, username, password string) (*DockerHubClient, error) {
	b, _ := json.Marshal(map[string]string{
		"username": username,
		"password": password,
	})
	req, _ := http.NewRequestWithContext(ctx,
		http.MethodPost,
		"https://hub.docker.com/v2/users/login/",
		bytes.NewReader(b),
	)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("login failed, got status %s", resp.Status)
	}

	var out struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}

	return &DockerHubClient{token: out.Token}, nil
}

func (c *DockerHubClient) UpsertRepo(ctx context.Context, dir, name string) error {
	exist, err := c.repoExists(ctx, name)
	if err != nil {
		return err
	}
	if exist {
		return nil
	}
	return c.createRepo(ctx, name, fmt.Sprintf("Generated from https://github.com/ctfer-io/recipes/blob/main/%s", dir))
}

func (c *DockerHubClient) repoExists(ctx context.Context, name string) (bool, error) {
	req, _ := http.NewRequestWithContext(ctx,
		http.MethodGet,
		fmt.Sprintf("https://hub.docker.com/v2/repositories/ctferio/%s/", name),
		nil,
	)
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return true, nil
	}
	if resp.StatusCode == http.StatusNotFound {
		return false, nil
	}

	return false, fmt.Errorf("unexpected status: %d", resp.StatusCode)
}

func (c *DockerHubClient) createRepo(ctx context.Context, name, description string) error {
	b, _ := json.Marshal(map[string]any{
		"registry":    "docker",
		"namespace":   "ctferio",
		"is_private":  false,
		"name":        name,
		"description": description,
	})
	req, _ := http.NewRequestWithContext(ctx,
		"POST",
		"https://hub.docker.com/v2/repositories/",
		bytes.NewReader(b),
	)
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to create repo: %s", body)
	}

	return nil
}
