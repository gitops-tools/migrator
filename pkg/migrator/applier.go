package migrator

import (
	"fmt"

	jsonpatch "github.com/evanphx/json-patch/v5"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const (
	mergePatchType = "application/merge-patch+json"
	jsonPatchType  = "application/json-patch+json"
)

// ApplyPatches applies a set of "patches" to a resource.
//
// A copy of the resource is returned with the patches applied.
func ApplyPatches(obj *unstructured.Unstructured, patches []Patch) (*unstructured.Unstructured, error) {
	objCopy := obj.DeepCopy() // DeepCopy requires a pointer to obj
	var err error
	for _, patch := range patches {
		switch patch.Type {
		case jsonPatchType:
			objCopy, err = applyJSONPatch(objCopy, patch.Change)
			if err != nil {
				return nil, err
			}
		case mergePatchType:
			objCopy, err = applyMergePatch(objCopy, patch.Change)
			if err != nil {
				return nil, err
			}
		default:
			return nil, fmt.Errorf("unknown patch type: %s", patch.Type)
		}
	}

	return objCopy, nil
}

func applyJSONPatch(obj *unstructured.Unstructured, change string) (*unstructured.Unstructured, error) {
	patch, err := jsonpatch.DecodePatch([]byte(change))
	if err != nil {
		return nil, fmt.Errorf("decoding patch: %w", err)
	}

	return applyPatch(obj, func(b []byte) ([]byte, error) {
		return patch.Apply(b)
	})
}

func applyMergePatch(obj *unstructured.Unstructured, change string) (*unstructured.Unstructured, error) {
	return applyPatch(obj, func(b []byte) ([]byte, error) {
		return jsonpatch.MergePatch(b, []byte(change))
	})
}

type patchApplier func([]byte) ([]byte, error)

func applyPatch(obj *unstructured.Unstructured, f patchApplier) (*unstructured.Unstructured, error) {
	b, err := obj.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("marshalling resource to JSON for patching: %w", err)
	}

	patched, err := f(b)
	if err != nil {
		// TODO
		return nil, err
	}

	if err := obj.UnmarshalJSON(patched); err != nil {
		return nil, fmt.Errorf("unmarshalling resource to JSON after patching: %w", err)
	}

	return obj, nil
}
