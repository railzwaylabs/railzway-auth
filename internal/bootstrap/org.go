package bootstrap

import (
	"context"
	"fmt"
	"time"

	"github.com/bwmarrin/snowflake"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/smallbiznis/railzway-auth/internal/config"
	"github.com/smallbiznis/railzway-auth/internal/domain"
	"github.com/smallbiznis/railzway-auth/internal/repository"
)

// EnsureOrg creates the default organization if it doesn't exist.
func EnsureOrg(lc fx.Lifecycle, cfg config.Config, orgs repository.OrgRepository, node *snowflake.Node, logger *zap.Logger) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return ensureOrg(ctx, cfg, orgs, node, logger)
		},
	})
}

func ensureOrg(ctx context.Context, cfg config.Config, orgs repository.OrgRepository, node *snowflake.Node, logger *zap.Logger) error {
	// Check count
	count, err := orgs.Count(ctx)
	if err != nil {
		return fmt.Errorf("bootstrap org count: %w", err)
	}

	if count > 0 {
		logger.Info("bootstrap org check: organizations exist, skipping default creation", zap.Int64("count", count))
		return nil
	}

	// Create default org
	orgID := int64(2016070718164307968)
	newOrg := domain.Org{
		ID:        orgID,
		Name:      "Default Organization",
		Slug:      fmt.Sprintf("org-%d", orgID),
		Status:    "ACTIVE",
		Type:      "organization",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	created, err := orgs.Create(ctx, newOrg)
	if err != nil {
		return fmt.Errorf("bootstrap create org: %w", err)
	}

	logger.Info("bootstrap org created",
		zap.Int64("org_id", created.ID),
		zap.String("slug", created.Slug),
	)

	return nil
}
