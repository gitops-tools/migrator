package migrator

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestApplyPatches(t *testing.T) {
	cm := newConfigMap()

	patches := Patch{
		JSONPatches: []Patch6902{
			{
				Op:    "replace",
				Path:  "/data/testing",
				Value: "new-value",
			},
		},
	}

	updated, err := ApplyPatches(toUnstructured(t, cm), patches)
	assert.NoError(t, err)

	want := &unstructured.Unstructured{
		Object: map[string]any{
			"apiVersion": "v1",
			"data": map[string]any{
				"testing": "new-value",
			},
			"kind": "ConfigMap",
			"metadata": map[string]any{
				"creationTimestamp": nil,
				"name":              "test-cm",
				"namespace":         "default",
			},
		},
	}
	if diff := cmp.Diff(want, updated); diff != "" {
		t.Fatalf("failed to apply migrations:\n%s", diff)
	}
}

func TestApplyPatches_invalid_patch(t *testing.T) {
	// cm := newConfigMap()

	// patches := []Patch{
	// 	{
	// 		Type:   "application/json-patch+json",
	// 		Change: `[{"op":"replace","path":"/data/testing","value":"new-value"}]`,
	// 	},
	// }

	// updated, err := ApplyPatches(toUnstructured(t, cm), patches)
	// assert.NoError(t, err)

	// want := &unstructured.Unstructured{
	// 	Object: map[string]any{
	// 		"apiVersion": "v1",
	// 		"data": map[string]any{
	// 			"testing": "new-value",
	// 		},
	// 		"kind": "ConfigMap",
	// 		"metadata": map[string]any{
	// 			"creationTimestamp": nil,
	// 			"name":              "test-cm",
	// 			"namespace":         "default",
	// 		},
	// 	},
	// }
	// if diff := cmp.Diff(want, updated); diff != "" {
	// 	t.Fatalf("failed to apply migrations:\n%s", diff)
	// }

}

func TestApplyPatches_fail_to_patch(t *testing.T) {
	// cm := newConfigMap()

	// patches := []Patch{
	// 	{
	// 		Type:   "application/json-patch+json",
	// 		Change: `[{"op":"replace","path":"/data/1/testing","value":"new-value"}]`,
	// 	},
	// }

	// _, err := ApplyPatches(toUnstructured(t, cm), patches)
	// assert.ErrorContains(t, err, "replace operation does not apply: doc is missing path")
}

func newConfigMap(opts ...func(*corev1.ConfigMap)) *corev1.ConfigMap {
	cm := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-cm",
			Namespace: "default",
		},
		Data: map[string]string{
			"testing": "test",
		},
	}

	for _, o := range opts {
		o(cm)
	}

	return cm
}

func toUnstructured(t *testing.T, obj runtime.Object) *unstructured.Unstructured {
	t.Helper()
	raw, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	assert.NoError(t, err)

	return &unstructured.Unstructured{Object: raw}
}
