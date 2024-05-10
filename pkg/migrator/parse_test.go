package migrator

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

// Migration describes a change that is applied to a resource.
type Migration struct {
	Filename string
	Name     string `yaml:"name"`
}

// ParseMigrations parses all the yaml files in the migration directory.
func ParseMigrations(dir string) ([]Migration, error) {
	return nil, nil
}

func TestParseMigrations(t *testing.T) {
	migrations, err := ParseMigrations("testdata/simple")
	if err != nil {
		t.Fatal(err)
	}

	want := []Migration{
		{
			Name: "patch-broken-authconfig-secret-name",
		},
	}
	if diff := cmp.Diff(want, migrations); diff != "" {
		t.Fatalf("failed to parse migrations:\n%s", diff)
	}
}
