package migrator

import (
	"testing"

	"github.com/bigkevmcd/migrator/pkg/migrator/celpatch"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestApplyPatches_json_patch(t *testing.T) {
	cm := newConfigMap()

	patches := []Patch{
		{
			Type:   "application/json-patch+json",
			Change: `[{"op":"replace","path":"/data/testing","value":"new-value"}]`,
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
	cm := newConfigMap()

	patches := []Patch{
		{
			Type:   "application/json-patch+json",
			Change: `[{"op":"replace","path":"/data/testing","value":"new-value"}]`,
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

func TestApplyPatches_fail_to_patch(t *testing.T) {
	cm := newConfigMap()

	patches := []Patch{
		{
			Type:   "application/json-patch+json",
			Change: `[{"op":"replace","path":"/data/1/testing","value":"new-value"}]`,
		},
	}

	_, err := ApplyPatches(toUnstructured(t, cm), patches)
	assert.ErrorContains(t, err, "replace operation does not apply: doc is missing path")
}

func TestApplyPatches_merge_patch(t *testing.T) {
	svc := newService()

	patches := []Patch{
		{
			Type:   "application/merge-patch+json",
			Change: `{"spec":{"ports":[{"name":"http-80","port":80,"protocol":"TCP","targetPort":9080}]}}`,
		},
	}

	updated, err := ApplyPatches(toUnstructured(t, svc), patches)
	assert.NoError(t, err)

	want := &unstructured.Unstructured{
		Object: map[string]any{
			"apiVersion": "v1",
			"kind":       "Service",
			"metadata": map[string]any{
				"creationTimestamp": nil,
				"name":              "test-svc",
				"namespace":         "default",
			},
			"spec": map[string]any{
				"ports": []any{
					map[string]any{
						"name":       string("http-80"),
						"port":       int64(80),
						"protocol":   string("TCP"),
						"targetPort": int64(9080),
					},
				},
			},
			"status": map[string]any{"loadBalancer": map[string]any{}},
		},
	}
	if diff := cmp.Diff(want, updated); diff != "" {
		t.Fatalf("failed to apply migrations:\n%s", diff)
	}
}

func TestApplyPatches_cel_migration(t *testing.T) {
	cm := newConfigMap()

	patches := []Patch{
		{
			Type: "application/migrate-cel-patch",
			CEL: []celpatch.Change{
				{
					Key:      "data.testing",
					NewValue: "'this is migrated'",
				},
			},
		},
	}

	updated, err := ApplyPatches(toUnstructured(t, cm), patches)
	assert.NoError(t, err)

	want := &unstructured.Unstructured{
		Object: map[string]any{
			"apiVersion": "v1",
			"data": map[string]any{
				"testing": "this is migrated",
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

func newFakeClient(objs ...runtime.Object) client.Client {
	return fake.NewClientBuilder().
		WithRuntimeObjects(objs...).
		Build()
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

func newService(opts ...func(*corev1.Service)) *corev1.Service {
	svc := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-svc",
			Namespace: "default",
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:       "http-80",
					Protocol:   corev1.ProtocolTCP,
					Port:       80,
					TargetPort: intstr.FromInt(9376)},
			},
		},
	}

	for _, o := range opts {
		o(svc)
	}

	return svc
}

func toUnstructured(t *testing.T, obj runtime.Object) *unstructured.Unstructured {
	t.Helper()
	raw, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	assert.NoError(t, err)

	return &unstructured.Unstructured{Object: raw}
}
