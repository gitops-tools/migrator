package celpatch

import (
	"testing"

	"github.com/bigkevmcd/migrator/pkg/migrator/celpatch/ext"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/json"
)

func TestApplyChanges(t *testing.T) {
	cm := newConfigMap()
	changes := []Change{
		{
			Key:      "data.testing",
			NewValue: "'this is migrated'",
		},
		{
			Key:      "data.newKey",
			NewValue: "'This is added'",
		},
		{
			Key:      "data.otherKey",
			NewValue: "resource.data.tested",
		},
		{
			Key:      "data.personID",
			NewValue: "ldap.lookup('testuser@example.com').guid",
		},
	}

	updated, err := ApplyChanges(testMarshalToJSON(t, cm), changes, WithCELLib(ext.Demo()))
	assert.NoError(t, err)

	want := &unstructured.Unstructured{
		Object: map[string]any{
			"apiVersion": "v1",
			"data": map[string]any{
				"testing":  "this is migrated",
				"newKey":   "This is added",
				"otherKey": "this-value",
				"tested":   "this-value",
				"personID": "27f7c407-99d7-4c5a-8ebd-206a5c2e3f3d",
			},
			"kind": "ConfigMap",
			"metadata": map[string]any{
				"creationTimestamp": nil,
				"name":              "test-cm",
				"namespace":         "default",
			},
		},
	}

	obj := &unstructured.Unstructured{}
	if err := obj.UnmarshalJSON(updated); err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(want, obj); diff != "" {
		t.Fatalf("failed to apply migrations:\n%s", diff)
	}
}

func TestApplyChanges_migrate_user(t *testing.T) {
	user := &unstructured.Unstructured{
		Object: map[string]any{
			"apiVersion":  "management.cattle.io/v3",
			"description": "",
			"displayName": "System account for Cluster local",
			"kind":        "User",
			"metadata": map[string]any{
				"annotations": map[string]any{
					"authz.management.cattle.io/creator-role-bindings":      "{\"created\":[\"user\"],\"required\":[\"user\"]}",
					"lifecycle.cattle.io/create.mgmt-auth-users-controller": "true",
				},
				"creationTimestamp": "2024-10-02T09:30:38Z",
				"finalizers": []any{
					"controller.cattle.io/mgmt-auth-users-controller",
				},
				"generation": int64(3),
				"labels": map[string]any{
					"EDSN6T35DKT2UBRCDTHM2R0": "hashed-principal-name",
					"cattle.io/creator":       "norman",
				},
				"name":            "u-b4qkhsnliz",
				"resourceVersion": "3201",
				"uid":             "87b1367e-b440-4dda-801b-662209926c6d",
			},
			"principalIds": []string{
				"system://local",
				"ldap_user://cn=test,dc=example,dc=com",
			},
		},
	}

	changes := []Change{
		{
			Key:      "principalIds.1",
			NewValue: "'this is a test'",
		},
	}

	updated, err := ApplyChanges(testMarshalToJSON(t, user), changes, WithCELLib(ext.Demo()))
	assert.NoError(t, err)

	want := &unstructured.Unstructured{
		Object: map[string]any{
			"apiVersion":  "management.cattle.io/v3",
			"description": "",
			"displayName": "System account for Cluster local",
			"kind":        "User",
			"metadata": map[string]any{
				"annotations": map[string]any{
					"authz.management.cattle.io/creator-role-bindings":      "{\"created\":[\"user\"],\"required\":[\"user\"]}",
					"lifecycle.cattle.io/create.mgmt-auth-users-controller": "true",
				},
				"creationTimestamp": "2024-10-02T09:30:38Z",
				"finalizers": []any{
					"controller.cattle.io/mgmt-auth-users-controller",
				},
				"generation": int64(3),
				"labels": map[string]any{
					"EDSN6T35DKT2UBRCDTHM2R0": "hashed-principal-name",
					"cattle.io/creator":       "norman",
				},
				"name":            "u-b4qkhsnliz",
				"resourceVersion": "3201",
				"uid":             "87b1367e-b440-4dda-801b-662209926c6d",
			},
			"principalIds": []string{
				"system://local",
				"local://u-b4qkhsnliz",
			},
		},
	}

	obj := &unstructured.Unstructured{}
	if err := obj.UnmarshalJSON(updated); err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(want, obj); diff != "" {
		t.Fatalf("failed to apply migrations:\n%s", diff)
	}
}

func TestApplyChanges_errors(t *testing.T) {
	errTests := []struct {
		name    string
		changes []Change
		wantMsg string
	}{
		{
			name: "invalid new value",
			changes: []Change{
				{
					Key:      "data.testing",
					NewValue: "this is migrated",
				},
			},
			wantMsg: `failed to parse expression "this is migrated"`,
		},
		{
			name: "invalid key",
			changes: []Change{
				{
					Key:      "",
					NewValue: "'testing'",
				},
			},
			wantMsg: "path cannot be empty",
		},
		{
			name: "invalid expression for new value",
			changes: []Change{
				{
					Key:      "data.testing",
					NewValue: "this.test",
				},
			},
			wantMsg: "expression this.test check failed",
		},
		{
			name: "non-string value",
			changes: []Change{
				{
					Key:      "data.testing",
					NewValue: "52",
				},
			},
			wantMsg: `expression "52" did not evaluate to a string`,
		},
	}

	for _, tt := range errTests {
		t.Run(tt.name, func(t *testing.T) {
			cm := newConfigMap()

			_, err := ApplyChanges(testMarshalToJSON(t, cm), tt.changes)
			assert.ErrorContains(t, err, tt.wantMsg)
		})
	}

}

// TODO: Migrate to test directory!!
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
			"tested":  "this-value",
		},
	}

	for _, o := range opts {
		o(cm)
	}

	return cm
}

func testMarshalToJSON(t *testing.T, o runtime.Object) []byte {
	t.Helper()
	b, err := json.Marshal(o)
	if err != nil {
		t.Fatal(err)
	}

	return b
}
