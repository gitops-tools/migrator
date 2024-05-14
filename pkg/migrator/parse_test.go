package migrator

import (
	"testing"

	"sigs.k8s.io/kustomize/v3/pkg/gvk"
	"sigs.k8s.io/kustomize/v3/pkg/types"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
)

func TestParseDirectory_cluster_resource(t *testing.T) {
	migrations, err := ParseDirectory("testdata/clusterresource")
	if err != nil {
		t.Fatal(err)
	}

	want := []Migration{
		{
			Name:     "patch-broken-authconfig-secret-name",
			Filename: "testdata/clusterresource/patch_secret_name.yaml",
			Target: types.PatchTarget{
				Gvk: gvk.Gvk{
					Group:   "management.cattle.io",
					Version: "v3",
					Kind:    "AuthConfig",
				},
				Namespace: "",
				Name:      "shibboleth",
			},
			Up: []Patch{
				{
					Type:   "application/json-patch+json",
					Change: `[{"op":"replace","path":"/openLdapConfig/serviceAccountPassword","value":"cattle-global-data:shibbolethconfig-serviceaccountpassword"}]`,
				},
			},
			Down: []Patch{
				{
					Type:   "application/json-patch+json",
					Change: `[{"op":"replace","path":"/openLdapConfig/serviceAccountPassword","value":"cattle-global-data:shibbolethconfig-serviceAccountPassword"}]`,
				},
			},
		},
	}
	if diff := cmp.Diff(want, migrations); diff != "" {
		t.Fatalf("failed to parse migrations:\n%s", diff)
	}
}

func TestParseDirectory(t *testing.T) {
	migrations, err := ParseDirectory("testdata/simple")
	if err != nil {
		t.Fatal(err)
	}

	want := []Migration{
		{
			Name:     "migrate-service",
			Filename: "testdata/simple/migrate_service.yaml",
			Target: types.PatchTarget{
				Gvk: gvk.Gvk{
					Group:   "",
					Version: "v1",
					Kind:    "Service",
				},
				Namespace: "default",
				Name:      "test-service",
			},
			Up: []Patch{
				{
					Type:   "application/json-patch+json",
					Change: `[{"op":"replace","path":"/spec/ports/0/targetPort","value":9371}]`,
				},
			},
		},
	}
	if diff := cmp.Diff(want, migrations); diff != "" {
		t.Fatalf("failed to parse migrations:\n%s", diff)
	}
}

func TestParseDirectory_missing_dir(t *testing.T) {
	_, err := ParseDirectory("testdata/unknown")
	assert.ErrorContains(t, err, "reading directory testdata/unknown")
}

func TestParseDirectory_bad_file(t *testing.T) {
	_, err := ParseDirectory("testdata/symlink")
	assert.ErrorContains(t, err, "parsing migration testdata/symlink/test.yaml: open testdata/symlink/test.yaml")
}

func TestParseDirectory_invalid_yaml(t *testing.T) {
	_, err := ParseDirectory("testdata/bad_yaml")
	assert.ErrorContains(t, err, "error converting YAML to JSON")
}

func TestParseDirectory_name_ordering(t *testing.T) {
	// ParseDirectory _currently_ uses os.ReadDir which sorts on name.
	migrations, err := ParseDirectory("testdata/ordered")
	if err != nil {
		t.Fatal(err)
	}

	var migrationNames []string
	for _, m := range migrations {
		migrationNames = append(migrationNames, m.Filename)
	}

	want := []string{
		"testdata/ordered/01_patch_secret_name.yaml",
		"testdata/ordered/02_patch_secret_name.yaml",
		"testdata/ordered/03_patch_secret_name.yaml",
	}
	if diff := cmp.Diff(want, migrationNames); diff != "" {
		t.Fatalf("failed to parse migrations:\n%s", diff)
	}
}
