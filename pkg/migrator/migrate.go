package migrator

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// MigrateUp executes the migrations forward.
// TODO: option for batching!
// TODO: option for dry-run (do not do the patch).
func MigrateUp(ctx context.Context, kubeClient client.Client, migrations []Migration) error {
	return migrate(ctx, kubeClient, migrations, func(m Migration) []Patch {
		return m.Up
	})
}

// MigrateUp executes the migrations down.
func MigrateDown(ctx context.Context, kubeClient client.Client, migrations []Migration) error {
	return migrate(ctx, kubeClient, migrations, func(m Migration) []Patch {
		return m.Down
	})
}

func migrate(ctx context.Context, kubeClient client.Client, migrations []Migration, f func(Migration) []Patch) error {
	for _, migration := range migrations {
		toMigrate, err := resourcesToMigrate(ctx, kubeClient, migration)
		if err != nil {
			return err
		}

		for _, resource := range toMigrate {
			updated, err := ApplyPatches(&resource, f(migration))
			if err != nil {
				return err
			}

			if err := kubeClient.Patch(ctx, updated, client.MergeFrom(&resource)); err != nil {
				// TODO
				return err
			}
		}
	}

	return nil
}

func resourcesToMigrate(ctx context.Context, kubeClient client.Reader, migration Migration) ([]unstructured.Unstructured, error) {
	target := migration.TargetObjectKey()

	if target.Name != "" {
		return singleResource(ctx, kubeClient, target, migration)
	}

	return multiResources(ctx, kubeClient, target, migration)
}

func singleResource(ctx context.Context, kubeClient client.Reader, target client.ObjectKey, migration Migration) ([]unstructured.Unstructured, error) {
	u := unstructured.Unstructured{}
	u.SetGroupVersionKind(migration.TargetGroupVersionKind())

	if err := kubeClient.Get(ctx, target, &u); err != nil {
		return nil, fmt.Errorf("getting migration target %s %s: %w", u.GetKind(), target, err)
	}

	return []unstructured.Unstructured{u}, nil
}

func multiResources(ctx context.Context, kubeClient client.Reader, target client.ObjectKey, migration Migration) ([]unstructured.Unstructured, error) {
	ul := unstructured.UnstructuredList{}
	ul.SetGroupVersionKind(migration.TargetGroupVersionKind())

	if err := kubeClient.List(ctx, &ul); err != nil {
		return nil, fmt.Errorf("getting migration targets %s %s: %w", ul.GetKind(), target, err)
	}

	return ul.Items, nil
}
