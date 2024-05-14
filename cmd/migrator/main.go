package main

import (
	"github.com/bigkevmcd/migrator/pkg/migrator"
	"github.com/spf13/cobra"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

func newRootCmd() *cobra.Command {
	var migrationsPath string

	cmd := cobra.Command{
		Use: "migrator",
		RunE: func(cmd *cobra.Command, args []string) error {

			parsed, err := migrator.ParseDirectory(migrationsPath)
			if err != nil {
				return err
			}

			cfg, err := config.GetConfig()
			if err != nil {
				return err
			}

			kubeClient, err := client.New(cfg, client.Options{})
			if err != nil {
				return err
			}

			return migrator.Migrate(cmd.Context(), kubeClient, parsed)
		},
	}

	cmd.Flags().StringVar(&migrationsPath, "migrations-dir", "", "Path to migrations to apply")
	cobra.CheckErr(cmd.MarkFlagRequired("migrations-dir"))

	return &cmd
}

func main() {
	cobra.CheckErr(newRootCmd().Execute())
}
