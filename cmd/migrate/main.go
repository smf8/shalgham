package migrate

import (
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/source/file" // Imported for its side effects
	"github.com/sirupsen/logrus"
	"github.com/smf8/shalgham/config"
	"github.com/smf8/shalgham/postgres"

	psql "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/spf13/cobra"
)

const (
	flagPath = "path"
)

func main(path string, cfg config.Postgres) error {
	db := postgres.WithRetry(postgres.Create, cfg)

	defer func() {
		if err := db.Close(); err != nil {
			logrus.Errorf("postgres connection close error: %s", err.Error())
		}
	}()

	driver, err := psql.WithInstance(db.DB(), &psql.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance("file://"+path, cfg.DBName, driver)
	if err != nil {
		return err
	}

	if err := m.Up(); err == migrate.ErrNoChange {
		logrus.Info("no change detected. All migrations have already been applied!")
		return nil
	} else if err != nil {
		return err
	}

	return nil
}

// Register migrate command.
func Register(root *cobra.Command, cfg config.Config) {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Provides DB migration functionality",

		// PreRunE does some validation, parsing, etc. Then populates migrationPath
		PreRunE: func(cmd *cobra.Command, args []string) error {
			path, err := cmd.Flags().GetString(flagPath)
			if err != nil {
				return fmt.Errorf("error parsing %s flag: %w", flagPath, err)
			}
			if path == "" {
				return fmt.Errorf("%s flag is required", flagPath)
			}

			return nil
		},

		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := cmd.Flags().GetString(flagPath)
			if err != nil {
				return err
			}

			if err := main(path, cfg.Postgres); err != nil {
				return err
			}

			cmd.Println("migrations ran successfully")

			return nil
		},
	}

	cmd.Flags().StringP(flagPath, "p", "", "migration folder path")

	root.AddCommand(cmd)
}
