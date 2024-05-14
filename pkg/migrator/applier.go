package migrator

import (
	"fmt"

	jsonpatch "github.com/evanphx/json-patch/v5"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// ApplyPatches applies a set of "patches" to a resource.
//
// A copy of the resource is returned with the patches applied.
func ApplyPatches(obj *unstructured.Unstructured, patches []Patch) (*unstructured.Unstructured, error) {
	objCopy := obj.DeepCopy()
	for _, patch := range patches {
		patch, err := jsonpatch.DecodePatch([]byte(patch.Change))
		if err != nil {
			return nil, fmt.Errorf("decoding patch: %w", err)
		}

		b, err := objCopy.MarshalJSON()
		if err != nil {
			return nil, fmt.Errorf("marshalling resource to JSON for patching: %s", err)
		}

		patched, err := patch.Apply(b)
		if err != nil {
			// TODO
			return nil, err
		}

		if err := objCopy.UnmarshalJSON(patched); err != nil {
			// TODO
			return nil, err
		}
	}

	return objCopy, nil
}
