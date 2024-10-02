package migrator

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/kustomize/v3/pkg/gvk"
	"sigs.k8s.io/kustomize/v3/pkg/types"
)

func TestMigrateChanges(t *testing.T) {
	migrationTests := []struct {
		name       string
		migrations []Migration
		want       []MigrationChange
	}{
		{
			name: "json-patch",
			migrations: []Migration{
				{
					Name:     "patch-service",
					Filename: "testdata/simple.yaml",
					Target: types.PatchTarget{
						Gvk: gvk.Gvk{
							Group:   "",
							Version: "v1",
							Kind:    "Service",
						},
						Namespace: "default",
						Name:      "test-svc",
					},
					Up: []Patch{
						{
							Type:   "application/json-patch+json",
							Change: `[{"op":"replace","path":"/spec/ports/0/port","value":81},{"op":"add","path":"/spec/selector","value":{"app":"test"}}]`,
						},
					},
				},
			},
			want: []MigrationChange{
				{
					Patch: PatchedResource{
						Name: client.ObjectKey{
							Name:      "test-svc",
							Namespace: "default",
						},
						GVK: schema.GroupVersionKind{
							Version: "v1",
							Kind:    "Service",
						},
						Patch: `{"spec":{"ports":[{"name":"http-80","port":81,"protocol":"TCP","targetPort":9376}],"selector":{"app":"test"}}}`,
					},
				},
			},
		},
	}

	for _, tt := range migrationTests {
		t.Run(tt.name, func(t *testing.T) {
			fc := fake.NewClientBuilder().WithObjects(newService()).Build()

			changes, err := CalculateChanges(context.TODO(), fc, tt.migrations)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(changes, tt.want); diff != "" {
				t.Fatalf("failed to calculate changes:\n%s", diff)
			}
		})
	}
}
