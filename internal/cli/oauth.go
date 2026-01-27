package cli

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"github.com/bwmarrin/snowflake"
	"github.com/spf13/cobra"

	"github.com/smallbiznis/railzway-auth/internal/domain"
	"github.com/smallbiznis/railzway-auth/internal/repository"
)

// Helper to init repos
func initRepos() (*repository.PostgresOAuthAppRepo, *repository.PostgresOAuthClientRepo, *snowflake.Node, error) {
	cfg, err := newConfig()
	if err != nil {
		return nil, nil, nil, err
	}
	pool, err := newPGXPool(nil, cfg)
	if err != nil {
		return nil, nil, nil, err
	}
	// Pool technically leaks if we don't return closer, but CLI ephemeral usage.
	// For better practice, we should ideally close it, but in `RunE` we can.
	// Refactoring to return pool.

	return repository.NewPostgresOAuthAppRepo(pool), repository.NewPostgresOAuthClientRepo(pool), nil, nil
}

// Better helper
func withRepos(f func(appRepo *repository.PostgresOAuthAppRepo, clientRepo *repository.PostgresOAuthClientRepo, snowflake *snowflake.Node) error) error {
	cfg, err := newConfig()
	if err != nil {
		return err
	}
	pool, err := newPGXPool(nil, cfg)
	if err != nil {
		return err
	}
	defer pool.Close()

	node, err := newSnowflake()
	if err != nil {
		return err
	}

	return f(repository.NewPostgresOAuthAppRepo(pool), repository.NewPostgresOAuthClientRepo(pool), node)
}

var oauthAppCmd = &cobra.Command{
	Use:   "oauth-app",
	Short: "Manage OAuth Applications",
}

var createOAuthAppCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an OAuth Application",
	RunE: func(cmd *cobra.Command, args []string) error {
		return withRepos(func(appRepo *repository.PostgresOAuthAppRepo, clientRepo *repository.PostgresOAuthClientRepo, node *snowflake.Node) error {
			name, _ := cmd.Flags().GetString("name")
			orgID, _ := cmd.Flags().GetInt64("org-id")
			appType, _ := cmd.Flags().GetString("type") // WEB, MOBILE, M2M

			if name == "" || orgID == 0 {
				return fmt.Errorf("name and org-id are required")
			}

			app := domain.OAuthApp{
				ID:           node.Generate().Int64(),
				OrgID:        orgID,
				Name:         name,
				Type:         appType,
				IsActive:     true,
				IsFirstParty: true,
			}

			created, err := appRepo.Create(context.Background(), app)
			if err != nil {
				return err
			}

			fmt.Printf("Created OAuth App: %s (ID: %d)\n", created.Name, created.ID)
			return nil
		})
	},
}

var oauthClientCmd = &cobra.Command{
	Use:   "oauth-client",
	Short: "Manage OAuth Clients",
}

var createOAuthClientCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an OAuth Client",
	RunE: func(cmd *cobra.Command, args []string) error {
		return withRepos(func(appRepo *repository.PostgresOAuthAppRepo, clientRepo *repository.PostgresOAuthClientRepo, node *snowflake.Node) error {
			orgID, _ := cmd.Flags().GetInt64("org-id")
			appID, _ := cmd.Flags().GetInt64("app-id")
			name, _ := cmd.Flags().GetString("name") // Use as client's logical name? client_ids are usually name-like or random.
			// The user said "oauth_client per app".
			// We need client_id and client_secret.
			
			if orgID == 0 || appID == 0 || name == "" {
				return fmt.Errorf("org-id, app-id, and name are required")
			}
			
			clientID := fmt.Sprintf("%s-%d", name, node.Generate().Int64())
			secret, _ := secureRandomString(32) // Reuse from oauth_service or dupe

			client := domain.OAuthClient{
				ID:             node.Generate().Int64(),
				OrgID:          orgID,
				AppID:          &appID,
				ClientID:       clientID,
				ClientSecret:   secret,
				Grants:         []string{"authorization_code", "refresh_token"},
				Scopes:         []string{"openid", "profile", "email"},
				RequireConsent: false,
			}

			created, err := clientRepo.UpsertClient(context.Background(), client)
			if err != nil {
				return err
			}

			fmt.Printf("Created OAuth Client:\nID: %s\nSecret: %s\nAppID: %d\n", created.ClientID, created.ClientSecret, *created.AppID)
			return nil
		})
	},
}

func init() {
	rootCmd.AddCommand(oauthAppCmd)
	oauthAppCmd.AddCommand(createOAuthAppCmd)

	createOAuthAppCmd.Flags().String("name", "", "Application Name")
	createOAuthAppCmd.Flags().Int64("org-id", 0, "Organization ID")
	createOAuthAppCmd.Flags().String("type", "WEB", "App Type (WEB, MOBILE, M2M)")
	createOAuthAppCmd.MarkFlagRequired("name")
	createOAuthAppCmd.MarkFlagRequired("org-id")

	rootCmd.AddCommand(oauthClientCmd)
	oauthClientCmd.AddCommand(createOAuthClientCmd)

	createOAuthClientCmd.Flags().String("name", "", "Client Name prefix")
	createOAuthClientCmd.Flags().Int64("org-id", 0, "Organization ID")
	createOAuthClientCmd.Flags().Int64("app-id", 0, "App ID this client belongs to")
	createOAuthClientCmd.MarkFlagRequired("name")
	createOAuthClientCmd.MarkFlagRequired("org-id")
	createOAuthClientCmd.MarkFlagRequired("app-id")
}

func secureRandomString(size int) (string, error) {
	if size <= 0 {
		size = 32
	}
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}
