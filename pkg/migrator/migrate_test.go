package migrator

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/kustomize/v3/pkg/gvk"
	"sigs.k8s.io/kustomize/v3/pkg/types"
)

func TestMigrate_merge_patch(t *testing.T) {
	// TODO: parse the Type and use a different type for merge patches!
	t.Skip()
}

func TestMigrateUp(t *testing.T) {
	migrations := []Migration{
		{
			Name:     "patch-service",
			Filename: "testdata/simple.yaml",
			Target: types.PatchTarget{
				Gvk: gvk.Gvk{
					Group:   "",
					Version: "v1",
					Kind:    "Service",
				},
				Namespace: "default",
				Name:      "testing",
			},
			Up: []Patch{
				{
					Type:   "application/json-patch+json",
					Change: `[{"op":"replace","path":"/spec/ports/0/port","value":81}]`,
				},
			},
		},
	}

	fc := fake.NewClientBuilder().WithObjects(createService()).Build()

	if err := MigrateUp(context.TODO(), fc, migrations); err != nil {
		t.Fatal(err)
	}

	var svc corev1.Service
	if err := fc.Get(context.TODO(), client.ObjectKey{Name: "testing", Namespace: "default"}, &svc); err != nil {
		t.Fatal(err)
	}

	want := []corev1.ServicePort{
		{
			Protocol:   "TCP",
			Port:       81,
			TargetPort: intstr.FromInt(9376),
		},
	}
	if diff := cmp.Diff(want, svc.Spec.Ports); diff != "" {
		t.Errorf("failed to migrate:\n%s", diff)
	}
}

func TestMigrateUp_missing_resource(t *testing.T) {
	migrations := []Migration{
		{
			Name:     "patch-service",
			Filename: "testdata/simple.yaml",
			Target: types.PatchTarget{
				Gvk: gvk.Gvk{
					Group:   "",
					Version: "v1",
					Kind:    "Service",
				},
				Namespace: "default",
				Name:      "testing",
			},
			Up: []Patch{
				{
					Type:   "application/json-patch+json",
					Change: `[{"op":"replace","path":"/spec/ports/0/port","value":81}]`,
				},
			},
		},
	}

	fc := fake.NewClientBuilder().Build()

	err := MigrateUp(context.TODO(), fc, migrations)
	assert.ErrorContains(t, err, "getting migration target Service default/testing: services \"testing\" not found")
}

func TestMigrateUp_bad_resource(t *testing.T) {
	migrations := []Migration{
		{
			Name:     "patch-service",
			Filename: "testdata/simple.yaml",
			Target: types.PatchTarget{
				Gvk: gvk.Gvk{
					Group:   "",
					Version: "v1",
					Kind:    "Service",
				},
				Namespace: "default",
				Name:      "testing",
			},
			Up: []Patch{
				{
					Type:   "application/json-patch+json",
					Change: `[{"op":"replace","path":"/spec/sports/0/port","value":81}]`,
				},
			},
		},
	}

	fc := fake.NewClientBuilder().WithObjects(createService()).Build()

	err := MigrateUp(context.TODO(), fc, migrations)
	assert.ErrorContains(t, err, "replace operation does not apply: doc is missing path: /spec/sports/0/port")
}

func TestMigrateDown(t *testing.T) {
	migrations := []Migration{
		{
			Name:     "patch-service",
			Filename: "testdata/simple.yaml",
			Target: types.PatchTarget{
				Gvk: gvk.Gvk{
					Group:   "",
					Version: "v1",
					Kind:    "Service",
				},
				Namespace: "default",
				Name:      "testing",
			},
			Up: []Patch{
				{
					Type:   "application/json-patch+json",
					Change: `[{"op":"replace","path":"/spec/ports/0/port","value":81}]`,
				},
			},
			Down: []Patch{
				{
					Type:   "application/json-patch+json",
					Change: `[{"op":"replace","path":"/spec/ports/0/port","value":80}]`,
				},
			},
		},
	}

	fc := fake.NewClientBuilder().WithObjects(createService()).Build()

	if err := MigrateUp(context.TODO(), fc, migrations); err != nil {
		t.Fatal(err)
	}

	if err := MigrateDown(context.TODO(), fc, migrations); err != nil {
		t.Fatal(err)
	}

	var svc corev1.Service
	if err := fc.Get(context.TODO(), client.ObjectKey{Name: "testing", Namespace: "default"}, &svc); err != nil {
		t.Fatal(err)
	}

	want := []corev1.ServicePort{
		{
			Protocol:   "TCP",
			Port:       80,
			TargetPort: intstr.FromInt(9376),
		},
	}
	if diff := cmp.Diff(want, svc.Spec.Ports); diff != "" {
		t.Errorf("failed to migrate:\n%s", diff)
	}
}

func createService() *corev1.Service {
	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testing",
			Namespace: "default",
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app.kubernetes.io/name": "MyApp",
			},
			Ports: []corev1.ServicePort{
				{
					Protocol:   "TCP",
					Port:       80,
					TargetPort: intstr.FromInt(9376),
				},
			},
		},
	}
}
