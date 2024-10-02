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

func TestMigrateUp(t *testing.T) {
	migrationTests := []struct {
		name       string
		migrations []Migration
		want       []corev1.ServicePort
	}{
		{
			name: "json-patch",
			migrations: []Migration{
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
						Name:      "test-svc",
					},
					Up: []Patch{
						{
							Type:   "application/json-patch+json",
							Change: `[{"op":"replace","path":"/spec/ports/0/port","value":81}]`,
						},
					},
				},
			},
			want: []corev1.ServicePort{
				{
					Protocol:   "TCP",
					Name:       "http-80",
					Port:       81,
					TargetPort: intstr.FromInt(9376),
				},
			},
		},
		{
			name: "merge-patch",
			migrations: []Migration{
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
						Name:      "test-svc",
					},
					Up: []Patch{
						{
							Type:   "application/merge-patch+json",
							Change: `{"spec":{"ports":[{"name":"http","port":80,"protocol":"TCP","targetPort":82}]}}`,
						},
					},
				},
			},
			want: []corev1.ServicePort{
				{
					Protocol:   "TCP",
					Name:       "http",
					Port:       80,
					TargetPort: intstr.FromInt(82),
				},
			},
		},
	}

	for _, tt := range migrationTests {
		t.Run(tt.name, func(t *testing.T) {
			fc := fake.NewClientBuilder().WithObjects(newService()).Build()

			if err := MigrateUp(context.TODO(), fc, tt.migrations); err != nil {
				t.Fatal(err)
			}

			var svc corev1.Service
			if err := fc.Get(context.TODO(), client.ObjectKey{
				Name: "test-svc", Namespace: "default"}, &svc); err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(tt.want, svc.Spec.Ports); diff != "" {
				t.Errorf("failed to migrate:\n%s", diff)
			}
		})
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
				Name:      "test-svc",
			},
			Up: []Patch{
				{
					Type:   "application/json-patch+json",
					Change: `[{"op":"replace","path":"/spec/sports/0/port","value":81}]`,
				},
			},
		},
	}

	fc := fake.NewClientBuilder().WithObjects(newService()).Build()

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
				Name:      "test-svc",
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

	fc := fake.NewClientBuilder().WithObjects(newService()).Build()

	if err := MigrateUp(context.TODO(), fc, migrations); err != nil {
		t.Fatal(err)
	}

	if err := MigrateDown(context.TODO(), fc, migrations); err != nil {
		t.Fatal(err)
	}

	var svc corev1.Service
	if err := fc.Get(context.TODO(), client.ObjectKey{Name: "test-svc", Namespace: "default"}, &svc); err != nil {
		t.Fatal(err)
	}

	want := []corev1.ServicePort{
		{
			Protocol:   "TCP",
			Port:       80,
			Name:       "http-80",
			TargetPort: intstr.FromInt(9376),
		},
	}
	if diff := cmp.Diff(want, svc.Spec.Ports); diff != "" {
		t.Errorf("failed to migrate:\n%s", diff)
	}
}

func TestMigrateUp_multiple_resources(t *testing.T) {
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
			},
			Up: []Patch{
				{
					Type:   "application/json-patch+json",
					Change: `[{"op":"replace","path":"/spec/ports/0/port","value":81}]`,
				},
			},
		},
	}

	fc := fake.NewClientBuilder().WithObjects(
		createService(withName("svc-1")),
		createService(withName("svc-2"))).Build()

	if err := MigrateUp(context.TODO(), fc, migrations); err != nil {
		t.Fatal(err)
	}

	var svcList corev1.ServiceList
	if err := fc.List(context.TODO(), &svcList, client.InNamespace("default")); err != nil {
		t.Fatal(err)
	}

	want := [][]corev1.ServicePort{
		{
			{Protocol: "TCP",
				Port:       81,
				TargetPort: intstr.FromInt(9376),
			},
		},
		{
			{
				Protocol:   "TCP",
				Port:       81,
				TargetPort: intstr.FromInt(9376),
			},
		},
	}

	ports := collect(svcList.Items, func(s corev1.Service) []corev1.ServicePort {
		return s.Spec.Ports
	})

	if diff := cmp.Diff(want, ports); diff != "" {
		t.Errorf("failed to migrate:\n%s", diff)
	}
}

func withName(s string) func(*corev1.Service) {
	return func(svc *corev1.Service) {
		svc.SetName(s)
	}
}

func createService(opts ...func(*corev1.Service)) *corev1.Service {
	svc := &corev1.Service{
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

	for _, opt := range opts {
		opt(svc)
	}

	return svc
}

func collect[T any, R any](input []T, pred func(T) R) []R {
	result := []R{}

	for _, v := range input {
		result = append(result, pred(v))
	}

	return result
}
