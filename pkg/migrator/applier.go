package migrator

import (
	"encoding/json"
	"fmt"
	"log"

	jsonpatch "github.com/evanphx/json-patch"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// ApplyPatches applies a set of "patches" to a resource.
//
// A copy of the resource is returned with the patches applied.
func ApplyPatches(obj *unstructured.Unstructured, p Patch) (*unstructured.Unstructured, error) {
	patch := convertPatch(p)

	b, err := obj.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("marshalling resource to JSON for patching: %s", err)
	}

	patched, err := patch.Apply(b)
	if err != nil {
		// TODO
		return nil, fmt.Errorf("applying patches: %w", err)
	}

	if err := obj.UnmarshalJSON(patched); err != nil {
		// TODO
		return nil, err
	}

	return obj, nil
}

func convertPatch(in Patch) jsonpatch.Patch {
	var p jsonpatch.Patch
	for _, v := range in.JSONPatches {
		log.Printf("KEVIN!!! Appending %#v", v)
		op := jsonpatch.Operation{
			"op":    rawMessage(v.Op),
			"path":  rawMessage(v.Path),
			"value": rawMessage(v.Value),
		}
		p = append(p, op)

		vo := op["op"]
		log.Printf("KEVIN!!!! unmarshalling %s to a string", *vo)
		var vs string
		if err := json.Unmarshal(*vo, &vs); err != nil {
			log.Printf("did not unmarshal: %s", err)
		}

		log.Printf("KEVIN!!!! op = %#v, %s vs %s", p, op.Kind(), vs)

	}

	return p
}

func rawMessage(s string) *json.RawMessage {
	msg := json.RawMessage([]byte(s))

	return &msg
}
