package migrator

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"sigs.k8s.io/yaml"
)

// Migration describes a change that is applied to a resource.
type Migration struct {
	Filename string
	Name     string `json:"name"`
	Patch    Patch  `json:"patch"`
}

type Selector struct {
	// // ResId refers to a GVKN/Ns of a resource.
	// resid.ResId `json:",inline,omitempty" yaml:",inline,omitempty"`

	// AnnotationSelector is a string that follows the label selection expression
	// https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#api
	// It matches with the resource annotations.
	AnnotationSelector string `json:"annotationSelector,omitempty" yaml:"annotationSelector,omitempty"`

	// LabelSelector is a string that follows the label selection expression
	// https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#api
	// It matches with the resource labels.
	LabelSelector string `json:"labelSelector,omitempty" yaml:"labelSelector,omitempty"`
}

// ParseMigrations parses all the yaml files in the migration directory.
func ParseMigrations(dir string) ([]Migration, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("reading directory %s: %w", dir, err)
	}

	var migrations []Migration
	for _, file := range files {
		migration, err := readYAML(filepath.Join(dir, file.Name()))
		if err != nil {
			// TODO
			return nil, err
		}
		migrations = append(migrations, *migration)
	}

	return migrations, nil
}

func readYAML(filename string) (*Migration, error) {
	f, err := os.Open(filename)
	if err != nil {
		// TODO: fix
		return nil, err
	}

	defer f.Close()

	decoder := yaml.NewDecoder(f)
	var migration Migration
	if err := decoder.Decode(&migration); err != nil {
		// TODO: fix
		return nil, err
	}

	migration.Filename = filename

	return &migration, nil
}

func TestParseMigrations(t *testing.T) {
	migrations, err := ParseMigrations("testdata/simple")
	if err != nil {
		t.Fatal(err)
	}

	want := []Migration{
		{
			Name:     "patch-broken-authconfig-secret-name",
			Filename: "testdata/simple/patch_secret_name.yaml",
		},
	}
	if diff := cmp.Diff(want, migrations); diff != "" {
		t.Fatalf("failed to parse migrations:\n%s", diff)
	}
}
