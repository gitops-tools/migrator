package migrator

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// MigrateUp executes the migrations forward.
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
		u := &unstructured.Unstructured{}
		u.SetGroupVersionKind(migration.TargetGroupVersionKind())

		if err := kubeClient.Get(ctx, migration.TargetObjectKey(), u); err != nil {
			return fmt.Errorf("getting migration target %s %s: %w", u.GetKind(), migration.TargetObjectKey(), err)
		}

		updated, err := ApplyPatches(u, f(migration))
		if err != nil {
			return err
		}

		if err := kubeClient.Patch(ctx, updated, client.MergeFrom(u)); err != nil {
			// TODO
			return err
		}
	}

	return nil
}
