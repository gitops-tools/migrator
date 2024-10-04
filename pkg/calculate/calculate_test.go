package calculate

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestMigrateChanges(t *testing.T) {
	migrationTests := []struct {
		name      string
		migration Migration
		want      []*MigrationChange
	}{
		{
			migration: testMigrator{},
			want: []*MigrationChange{
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
						Patch: `{"spec":{"selector":{"app":"test"}}}`,
					},
				},
			},
		},
	}

	for _, tt := range migrationTests {
		t.Run(tt.name, func(t *testing.T) {
			fc := fake.NewClientBuilder().WithObjects(newService()).Build()

			changes, err := tt.migration.Migrate(context.TODO(), fc)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(tt.want, changes); diff != "" {
				t.Fatalf("failed to calculate changes:\n%s", diff)
			}
		})
	}
}

type testMigrator struct {
}

// This should have a MigrationOptions with batch size etc.
// We need to be careful here restarting the query with a batch-size could
// result in the the same X items being returned and never progressing the
// batch.

func (m testMigrator) Migrate(ctx context.Context, reader client.Reader) ([]*MigrationChange, error) {
	ul := unstructured.UnstructuredList{}
	ul.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "",
		Version: "v1",
		Kind:    "Service",
	})

	if err := reader.List(ctx, &ul); err != nil {
		return nil, err
	}

	var changes []*MigrationChange
	for _, original := range ul.Items {
		change, err := createChange(original, func(updated *unstructured.Unstructured) error {
			return unstructured.SetNestedMap(updated.UnstructuredContent(), map[string]any{"app": "test"}, "spec", "selector")
		})
		// TODO: multi-error!
		if err != nil {
			return nil, err
		}

		changes = append(changes, change)
	}

	return changes, nil
}

func createChange(original unstructured.Unstructured, f func(*unstructured.Unstructured) error) (*MigrationChange, error) {
	updated := original.DeepCopy()

	if err := f(updated); err != nil {
		return nil, err
	}

	change, err := unstructuredDiff(*updated, original)
	if err != nil {
		return nil, err // TODO
	}
	return &MigrationChange{
		patchedResource(&original, change),
	}, nil
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
