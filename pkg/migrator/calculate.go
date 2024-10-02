package migrator

import (
	"context"

	jsonpatch "github.com/evanphx/json-patch"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// PatchedResource references an resource and the change to apply to that resource.
type PatchedResource struct {
	Name  client.ObjectKey
	GVK   schema.GroupVersionKind
	Patch string
}

// MigrationChange represents a patched resource as part of applying a
// migration.
type MigrationChange struct {
	Patch PatchedResource
}

// CalculateChanges takes a set of migrations and returns the changes
// that would be applied.
// TODO: Migrate from Kustomize gvk.Gvk
func CalculateChanges(ctx context.Context, kubeClient client.Reader, migrations []Migration) ([]MigrationChange, error) {
	changes := []MigrationChange{}

	for _, migration := range migrations {
		toMigrate, err := resourcesToMigrate(ctx, kubeClient, migration)
		if err != nil {
			return nil, err
		}

		for _, resource := range toMigrate {
			updated, err := ApplyPatches(&resource, migration.Up)
			if err != nil {
				return nil, err
			}

			change, err := unstructuredDiff(*updated, resource)
			if err != nil {
				return nil, err // TODO
			}
			changes = append(changes, MigrationChange{
				patchedResource(&resource, change),
			})
		}
	}

	return changes, nil
}

func patchedResource(obj client.Object, patch string) PatchedResource {
	return PatchedResource{
		Name:  client.ObjectKeyFromObject(obj),
		GVK:   obj.GetObjectKind().GroupVersionKind(),
		Patch: patch,
	}
}

func unstructuredDiff(a, b unstructured.Unstructured) (string, error) {
	aJSON, err := a.MarshalJSON()
	if err != nil {
		return "", err // TODO
	}
	bJSON, err := b.MarshalJSON()
	if err != nil {
		return "", err // TODO
	}

	patch, err := jsonpatch.CreateMergePatch(bJSON, aJSON)
	if err != nil {
		return "", err // TODO
	}

	return string(patch), nil
}
