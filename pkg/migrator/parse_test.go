package migrator

import (
	"testing"

	"sigs.k8s.io/kustomize/v3/pkg/gvk"
	"sigs.k8s.io/kustomize/v3/pkg/types"

	"github.com/google/go-cmp/cmp"
)

func TestParseMigrations(t *testing.T) {
	migrations, err := ParseMigrations("testdata/simple")
	if err != nil {
		t.Fatal(err)
	}

	want := []Migration{
		{
			Name:     "patch-broken-authconfig-secret-name",
			Filename: "testdata/simple/patch_secret_name.yaml",
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

func TestParseMigrations_missing_dir(t *testing.T) {
	t.Skip()
}

func TestParseMigrations_invalid_yaml(t *testing.T) {
	t.Skip()
}

func TestParseMigrations_name_ordering(t *testing.T) {
	t.Skip()
}

func TestParseMigrations_skip_non_yaml(t *testing.T) {
	t.Skip()
}
