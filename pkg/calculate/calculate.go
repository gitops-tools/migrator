package calculate

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

// Migration is an interface describing how to make changes.
type Migration interface {
	// TODO Add options for batching?
	Migrate(context.Context, client.Reader) ([]*MigrationChange, error)
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
