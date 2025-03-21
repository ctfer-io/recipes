package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
		ecoPath := filepath.Join(pwd, eco)
		entries, err := os.ReadDir(ecoPath)
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
				Main:   filepath.Join(ecoPath, e.Name()),
				Binary: id,
				Env:    []string{"CGO_ENABLED=0"},
				GOOS:   []string{"linux"},
				GOArch: []string{"amd64"},
			})
		}
	}

	// Generate goreleaser on-the-fly configuration
	relConf["build"] = be
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
