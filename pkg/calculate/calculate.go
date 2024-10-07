package calculate

import (
	"context"
	"fmt"

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

// Calculate calculates the patches for a set of migrations.
//
// This should accept options for batching etc.
func Calculate(ctx context.Context, k8sClient client.Reader, migration Migration) ([]*MigrationChange, error) {
	resources, err := migration.Resources(ctx, k8sClient)
	if err != nil {
		return nil, fmt.Errorf("failed to get resources for migration: %w", err)
	}

	var migrationChanges []*MigrationChange

	for _, resource := range resources {
		changes, err := migration.Migrate(ctx, resource)
		if err != nil {
			// TODO: multierr!
			return nil, err
		}
		migrationChanges = append(migrationChanges, changes...)
	}

	return migrationChanges, nil
}

// Migration is an interface describing how to make changes.
type Migration interface {
	// TODO Add options for batching?
	Resources(context.Context, client.Reader) ([]unstructured.Unstructured, error)
	Migrate(context.Context, unstructured.Unstructured) ([]*MigrationChange, error)
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
