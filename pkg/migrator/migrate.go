package migrator

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Migrate executes the migrations.
//
// TODO: direction!
func Migrate(ctx context.Context, kubeClient client.Client, migrations []Migration) error {
	for _, migration := range migrations {
		u := &unstructured.Unstructured{}
		u.SetGroupVersionKind(migration.TargetGroupVersionKind())

		if err := kubeClient.Get(ctx, migration.TargetObjectKey(), u); err != nil {
			return fmt.Errorf("getting migration target %s %s: %w", u.GetKind(), migration.TargetObjectKey(), err)
		}

		updated, err := ApplyPatches(u, migration.Up)
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
