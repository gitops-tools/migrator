package migrator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	apitypes "k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/kustomize/v3/pkg/types"
	"sigs.k8s.io/yaml"
)

// https://github.com/redhat-cop/patch-operator
// https://pkg.go.dev/k8s.io/apimachinery/pkg/types#PatchType

// Patch provides a generic description of the change to be applied to a
// resource.
type Patch struct {
	Type   apitypes.PatchType `json:"type,omitempty"`
	Change string             `json:"change,omitempty"`
}

// Migration describes a change that is applied to a resource.
type Migration struct {
	Filename string
	Name     string            `json:"name"`
	Target   types.PatchTarget `json:"target"`
	Up       []Patch           `json:"up"`
	Down     []Patch           `json:"down,omitempty"`
}

// ParseMigrations parses all the yaml files in the migration directory.
func ParseMigrations(dir string) ([]Migration, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("reading directory %s: %w", dir, err)
	}

	var migrations []Migration
	for _, name := range filterYAMLFiles(files) {
		migration, err := readYAML(filepath.Join(dir, name))
		if err != nil {
			// TODO
			return nil, err
		}
		migrations = append(migrations, *migration)
	}

	return migrations, nil
}

func filterYAMLFiles(entries []os.DirEntry) []string {
	var filtered []string
	for _, e := range entries {
		if name := e.Name(); strings.HasSuffix(name, ".yaml") || strings.HasSuffix(name, ".yml") {
			filtered = append(filtered, name)
		}
	}

	return filtered
}

func readYAML(filename string) (*Migration, error) {
	b, err := os.ReadFile(filename)
	if err != nil {
		// TODO: fix
		return nil, err
	}

	var migration Migration
	if err := yaml.Unmarshal(b, &migration); err != nil {
		// TODO: fix
		return nil, err
	}

	migration.Filename = filename

	return &migration, nil
}
