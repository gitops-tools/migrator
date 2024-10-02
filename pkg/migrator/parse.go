package migrator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bigkevmcd/migrator/pkg/migrator/celpatch"
	"k8s.io/apimachinery/pkg/runtime/schema"
	apitypes "k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
	CEL    []celpatch.Change  `json:"cel,omitempty"`
}

// Migration describes a change that is applied to a resource.
type Migration struct {
	Filename string
	Name     string            `json:"name"`
	Target   types.PatchTarget `json:"target"`
	Up       []Patch           `json:"up"`
	Down     []Patch           `json:"down,omitempty"`
}

// TargetGroupVersionKind returns the GVK for the Target as a GroupVersionKind.
func (m Migration) TargetGroupVersionKind() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   m.Target.Group,
		Version: m.Target.Version,
		Kind:    m.Target.Kind,
	}
}

// TargetObjectKey returns the Key for loading the Target resource.
func (m Migration) TargetObjectKey() client.ObjectKey {
	return client.ObjectKey{
		Name:      m.Target.Name,
		Namespace: m.Target.Namespace,
	}
}

// ParseDirectory parses all the yaml files in the migration directory.
func ParseDirectory(dir string) ([]Migration, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("reading directory %s: %w", dir, err)
	}

	var migrations []Migration
	for _, name := range filterYAMLFiles(files) {
		fullname := filepath.Join(dir, name)
		migration, err := readYAML(fullname)
		if err != nil {
			return nil, fmt.Errorf("parsing migration %s: %w", fullname, err)
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
		return nil, err
	}

	var migration Migration
	if err := yaml.Unmarshal(b, &migration); err != nil {
		return nil, fmt.Errorf("parsing YAML: %w", err)
	}

	migration.Filename = filename

	return &migration, nil
}
