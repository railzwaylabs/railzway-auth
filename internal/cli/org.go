package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/smallbiznis/railzway-auth/internal/domain"
)

var orgCmd = &cobra.Command{
	Use:   "org",
	Short: "Manage organizations",
}

var createOrgCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new organization",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := newConfig()
		if err != nil {
			return err
		}

		pool, err := newPGXPool(nil, cfg) // Lifecycle is nil, careful with cleanup, but for CLI 1-off it's ok-ish or use dummy
		if err != nil {
			return err
		}
		defer pool.Close()

		q := newQueries(pool)
		orgRepo := newOrgRepository(pool, q)
		// We can add simple ID generation if we want, or rely on DB defaults (if sequence).
		// Currently Org ID is int64 (Snowflake usually).
		snowflakeNode, err := newSnowflake()
		if err != nil {
			return err
		}

		name, _ := cmd.Flags().GetString("name")
		slug, _ := cmd.Flags().GetString("slug")
		
		id := snowflakeNode.Generate().Int64()

		org := domain.Org{
			ID:     id,
			Name:   name,
			Slug:   slug,
			Status: "ACTIVE",
			Type:   "organization", // Default
		}

		created, err := orgRepo.Create(context.Background(), org)
		if err != nil {
			return fmt.Errorf("create org: %w", err)
		}

		fmt.Printf("Created Org: %s (ID: %d, Slug: %s)\n", created.Name, created.ID, created.Slug)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(orgCmd)
	orgCmd.AddCommand(createOrgCmd)

	createOrgCmd.Flags().String("name", "", "Organization Name")
	createOrgCmd.Flags().String("slug", "", "Organization Slug")
	createOrgCmd.MarkFlagRequired("name")
	createOrgCmd.MarkFlagRequired("slug")
}
