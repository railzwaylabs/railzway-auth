package bootstrap

import (
	"context"
	"fmt"
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/jackc/pgx/v5"
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
	if cfg.DefaultOrgID == 0 {
		return fmt.Errorf("bootstrap org missing required config: DEFAULT_ORG")
	}

	// Check if org exists
	exists, err := orgs.GetOrg(ctx, cfg.DefaultOrgID)
	if err == nil {
		logger.Info("bootstrap org already exists", zap.Int64("org_id", exists.ID), zap.String("slug", exists.Slug))
		return nil
	}
	if err != nil && err != pgx.ErrNoRows {
		return fmt.Errorf("bootstrap org lookup: %w", err)
	}

	// Create org
	newOrg := domain.Org{
		ID:        cfg.DefaultOrgID,
		Name:      "Default Organization", // You might want this configurable, but for now hardcoded or derived is fine
		Slug:      fmt.Sprintf("org-%d", cfg.DefaultOrgID),
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
