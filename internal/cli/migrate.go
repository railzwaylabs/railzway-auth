package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Database migration management",
}

var migrateUpCmd = &cobra.Command{
	Use:   "up",
	Short: "Apply all up migrations",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := newConfig()
		if err != nil {
			return err
		}

		ctx := context.Background()
		pool, err := newPGXPool(nil, cfg)
		if err != nil {
			return err
		}
		defer pool.Close()

		// Allow overriding migration path via env or flag, default to /migrations for Docker
		migrationPath := os.Getenv("MIGRATION_PATH")
		if migrationPath == "" {
			migrationPath = "/migrations"
		}

		// Fallback for local dev if /migrations doesn't exist
		if _, err := os.Stat(migrationPath); os.IsNotExist(err) {
			cwd, _ := os.Getwd()
			localPath := filepath.Join(cwd, "sql", "migrations")
			if _, err := os.Stat(localPath); err == nil {
				migrationPath = localPath
			}
		}

		fmt.Printf("Looking for migrations in: %s\n", migrationPath)

		files, err := os.ReadDir(migrationPath)
		if err != nil {
			return fmt.Errorf("read migration dir: %w", err)
		}

		for _, file := range files {
			if file.IsDir() || !strings.HasSuffix(file.Name(), ".sql") {
				continue
			}

			fmt.Printf("Applying migration: %s\n", file.Name())
			content, err := os.ReadFile(filepath.Join(migrationPath, file.Name()))
			if err != nil {
				return fmt.Errorf("read file %s: %w", file.Name(), err)
			}

			if _, err := pool.Exec(ctx, string(content)); err != nil {
				return fmt.Errorf("archive execution %s: %w", file.Name(), err)
			}
		}

		fmt.Println("Migrations applied successfully.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(migrateCmd)
	migrateCmd.AddCommand(migrateUpCmd)
}
