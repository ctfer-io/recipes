package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

var (
	ecosystems = []string{
		"chall-manager",
	}
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] %s\n", err)
		os.Exit(1)
	}
}

func run() error {
	pwd, err := os.Getwd()
	if err != nil {
		return err
	}

	// Parse goreleaser YAML raw content
	rawConf, err := os.ReadFile(filepath.Join(pwd, ".goreleaser.yaml"))
	if err != nil {
		return err
	}
	var relConf map[string]any
	if err := yaml.Unmarshal(rawConf, &relConf); err != nil {
		return err
	}

	// Find entries to build
	be := []*BuildEntry{}
	for _, eco := range ecosystems {
		entries, err := os.ReadDir(eco)
		if err != nil {
			return err
		}
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}

			id := genID(eco, e.Name())
			be = append(be, &BuildEntry{
				ID:     id,
				Main:   filepath.Join(eco, e.Name()),
				Binary: id,
				Env:    []string{"CGO_ENABLED=0"},
				GOOS:   []string{"linux"},
				GOArch: []string{"amd64"},
			})
		}
	}

	// Generate goreleaser on-the-fly configuration
	relConf["builds"] = be
	rawConf, err = yaml.Marshal(relConf)
	if err != nil {
		return err
	}

	fmt.Printf("%s\n", rawConf)
	return nil
}

func genID(eco, name string) string {
	name = strings.ReplaceAll(name, ".", "-") // replace dots by dashes (avoid type confusion)
	return fmt.Sprintf("%s_%s", eco, name)
}

type BuildEntry struct {
	ID     string   `yaml:"id"`
	Main   string   `yaml:"main"`
	Binary string   `yaml:"binary"`
	Env    []string `yaml:"env"`
	GOOS   []string `yaml:"goos"`
	GOArch []string `yaml:"goarch"`
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
