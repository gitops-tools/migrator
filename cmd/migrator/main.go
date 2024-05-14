package main

import (
	"fmt"
	"strings"

	"github.com/bigkevmcd/migrator/pkg/migrator"
	"github.com/spf13/cobra"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

func newRootCmd() *cobra.Command {
	var (
		migrationsPath string
		direction      string
	)

	cmd := cobra.Command{
		Use: "migrator",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			direction = strings.ToLower(direction)
			if !(direction == "up" || direction == "down") {
				return fmt.Errorf("%s is not a valid migration direction", direction)
			}

			return nil
		},
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

			switch direction {
			case "up":
				return migrator.MigrateUp(cmd.Context(), kubeClient, parsed)
			case "down":
				return migrator.MigrateDown(cmd.Context(), kubeClient, parsed)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&migrationsPath, "migrations-dir", "", "Path to migrations to apply")
	cobra.CheckErr(cmd.MarkFlagRequired("migrations-dir"))

	cmd.Flags().StringVar(&direction, "direction", "up", "Direction - up or down")

	return &cmd
}

func main() {
	cobra.CheckErr(newRootCmd().Execute())
}
