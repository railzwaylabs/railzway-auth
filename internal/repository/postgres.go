package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/smallbiznis/railzway-auth/internal/domain"
	"github.com/smallbiznis/railzway-auth/sqlc"
)

// Compile-time interface assertions.
var (
	_ OrgRepository         = (*PostgresOrgRepo)(nil)
	_ UserRepository        = (*PostgresUserRepo)(nil)
	_ TokenRepository       = (*PostgresTokenRepo)(nil)
	_ CodeRepository        = (*PostgresCodeRepo)(nil)
	_ KeyRepository         = (*PostgresKeyRepo)(nil)
	_ OAuthClientRepository = (*PostgresOAuthClientRepo)(nil)
	_ OAuthAppRepository    = (*PostgresOAuthAppRepo)(nil)
)

// PostgresOrgRepo implements OrgRepository using sqlc.
type PostgresOrgRepo struct {
	q  *sqlc.Queries
	db *pgxpool.Pool
}

func NewPostgresOrgRepo(db *pgxpool.Pool, q *sqlc.Queries) *PostgresOrgRepo {
	return &PostgresOrgRepo{q: q, db: db}
}

func (r *PostgresOrgRepo) GetDomainByHost(ctx context.Context, host string) (domain.Domain, error) {
	row, err := r.q.GetDomainByHost(ctx, host)
	if err != nil {
		return domain.Domain{}, fmt.Errorf("get domain: %w", err)
	}
	// TODO(v1): rename tenant_id columns to org_id in DB schema.
	return domain.Domain{ID: row.ID, Host: row.Host, OrgID: row.TenantID}, nil
}

func (r *PostgresOrgRepo) GetOrg(ctx context.Context, orgID int64) (domain.Org, error) {
	row, err := r.q.GetTenant(ctx, orgID)
	if err != nil {
		return domain.Org{}, fmt.Errorf("get org: %w", err)
	}
	return mapOrgRow(row), nil
}

func (r *PostgresOrgRepo) Create(ctx context.Context, org domain.Org) (domain.Org, error) {
	const query = `
INSERT INTO tenants (id, type, name, slug, status, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
RETURNING id, type, name, code, slug, country_code, timezone, is_default, status, created_at, updated_at`

	var row sqlc.GetTenantRow
	err := r.db.QueryRow(ctx, query,
		org.ID,
		"organization", // Default type if not in domain.Org, assume structure
		org.Name,
		org.Slug,
		org.Status, // Assuming Active
	).Scan(
		&row.ID,
		&row.Type,
		&row.Name,
		&row.Code,
		&row.Slug,
		&row.CountryCode,
		&row.Timezone,
		&row.IsDefault,
		&row.Status,
		&row.CreatedAt,
		&row.UpdatedAt,
	)
	if err != nil {
		return domain.Org{}, fmt.Errorf("create org: %w", err)
	}
	return mapOrgRow(row), nil
}

func (r *PostgresOrgRepo) GetOrgBySlug(ctx context.Context, slug string) (domain.Org, error) {
	row, err := r.q.GetTenantBySlug(ctx, slug)
	if err != nil {
		return domain.Org{}, fmt.Errorf("get org by slug: %w", err)
	}
	return mapOrgRow(row), nil
}

func (r *PostgresOrgRepo) GetPrimaryDomain(ctx context.Context, orgID int64) (domain.Domain, error) {
	row, err := r.q.GetPrimaryDomain(ctx, orgID)
	if err != nil {
		return domain.Domain{}, fmt.Errorf("get primary domain: %w", err)
	}
	return domain.Domain{ID: row.ID, Host: row.Host, OrgID: row.TenantID}, nil
}

func (r *PostgresOrgRepo) GetBranding(ctx context.Context, orgID int64) (domain.Branding, error) {
	row, err := r.q.GetBranding(ctx, orgID)
	if err != nil {
		return domain.Branding{}, fmt.Errorf("get branding: %w", err)
	}
	return domain.Branding{OrgID: row.TenantID, LogoURL: &row.LogoURL, PrimaryColor: &row.PrimaryColor}, nil
}

func (r *PostgresOrgRepo) ListAuthProviders(ctx context.Context, orgID int64) ([]domain.AuthProvider, error) {
	rows, err := r.q.ListAuthProviders(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("list auth providers: %w", err)
	}
	providers := make([]domain.AuthProvider, 0, len(rows))
	for _, row := range rows {
		var configID *int64
		if row.ProviderConfigID.Valid {
			val := row.ProviderConfigID.Int64
			configID = &val
		}
		providers = append(providers, domain.AuthProvider{
			ID:               row.ID,
			OrgID:            row.TenantID,
			ProviderType:     row.ProviderType,
			ProviderConfigID: configID,
			IsActive:         row.IsActive,
			CreatedAt:        row.CreatedAt,
			UpdatedAt:        row.UpdatedAt,
		})
	}
	return providers, nil
}

func (r *PostgresOrgRepo) GetPasswordConfig(ctx context.Context, orgID int64) (domain.PasswordConfig, error) {
	row, err := r.q.GetPasswordConfig(ctx, orgID)
	if err != nil {
		return domain.PasswordConfig{}, fmt.Errorf("get password config: %w", err)
	}
	return domain.PasswordConfig{
		OrgID:                  row.TenantID,
		MinLength:              int(row.MinLength),
		RequireUppercase:       row.RequireUppercase,
		RequireNumber:          row.RequireNumber,
		RequireSymbol:          row.RequireSymbol,
		AllowSignup:            row.AllowSignup,
		AllowPasswordReset:     row.AllowPasswordReset,
		LockoutAttempts:        int(row.LockoutAttempts),
		LockoutDurationSeconds: int(row.LockoutDurationSeconds),
		CreatedAt:              row.CreatedAt,
		UpdatedAt:              row.UpdatedAt,
	}, nil
}

func (r *PostgresOrgRepo) GetOTPConfig(ctx context.Context, orgID int64) (domain.OTPConfig, error) {
	row, err := r.q.GetOTPConfig(ctx, orgID)
	if err != nil {
		return domain.OTPConfig{}, fmt.Errorf("get otp config: %w", err)
	}
	return domain.OTPConfig{
		OrgID:         row.TenantID,
		Channel:       row.Channel,
		Provider:      row.Provider,
		APIKey:        row.APIKey.String,
		Sender:        row.Sender.String,
		Template:      row.Template.String,
		ExpirySeconds: int(row.ExpirySeconds),
		CreatedAt:     row.CreatedAt,
		UpdatedAt:     row.UpdatedAt,
	}, nil
}

func (r *PostgresOrgRepo) ListOAuthIDPConfigs(ctx context.Context, orgID int64) ([]domain.OAuthIDPConfig, error) {
	rows, err := r.q.ListOAuthIDPConfigs(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("list oauth idps: %w", err)
	}
	res := make([]domain.OAuthIDPConfig, 0, len(rows))
	for _, row := range rows {
		var scopes []string
		if len(row.Scopes) > 0 {
			scopes = row.Scopes
		}
		extra := map[string]any{}
		if len(row.Extra) > 0 {
			_ = json.Unmarshal(row.Extra, &extra)
		}
		res = append(res, domain.OAuthIDPConfig{
			OrgID:            row.TenantID,
			Provider:         row.Provider,
			ClientID:         row.ClientID,
			ClientSecret:     row.ClientSecret,
			IssuerURL:        row.IssuerURL,
			AuthorizationURL: row.AuthorizationURL,
			TokenURL:         row.TokenURL,
			UserinfoURL:      row.UserinfoURL,
			JWKSURL:          row.JWKSURL,
			Scopes:           scopes,
			Extra:            extra,
			CreatedAt:        row.CreatedAt,
			UpdatedAt:        row.UpdatedAt,
		})
	}
	return res, nil
}

// PostgresUserRepo implements UserRepository.
type PostgresUserRepo struct {
	q  *sqlc.Queries
	db *pgxpool.Pool
}

func NewPostgresUserRepo(pool *pgxpool.Pool) *PostgresUserRepo {
	return &PostgresUserRepo{q: sqlc.New(pool), db: pool}
}

func (r *PostgresUserRepo) GetByEmail(ctx context.Context, orgID int64, email string) (domain.User, error) {
	row, err := r.q.GetUserByEmail(ctx, orgID, email)
	if err != nil {
		return domain.User{}, fmt.Errorf("get user: %w", err)
	}
	return mapUserRow(row), nil
}

func (r *PostgresUserRepo) GetByID(ctx context.Context, orgID, userID int64) (domain.User, error) {
	row, err := r.q.GetUserByID(ctx, orgID, userID)
	if err != nil {
		return domain.User{}, fmt.Errorf("get user by id: %w", err)
	}
	return mapUserRow(row), nil
}

const insertUserSQL = `INSERT INTO users (id, tenant_id, email, email_verified, password_hash, name, phone, phone_verified, avatar_url, status)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING id, tenant_id, email, email_verified, password_hash, name, phone, phone_verified, avatar_url, status, created_at, updated_at`

func (r *PostgresUserRepo) Create(ctx context.Context, user domain.User) (domain.User, error) {
	row := r.db.QueryRow(ctx, insertUserSQL,
		user.ID,
		user.OrgID,
		user.Email,
		user.EmailVerified,
		user.PasswordHash,
		user.Name,
		user.Phone,
		user.PhoneVerified,
		user.AvatarURL,
		user.Status,
	)

	var inserted sqlc.GetUserByEmailRow
	if err := row.Scan(
		&inserted.ID,
		&inserted.TenantID,
		&inserted.Email,
		&inserted.EmailVerified,
		&inserted.PasswordHash,
		&inserted.Name,
		&inserted.Phone,
		&inserted.PhoneVerified,
		&inserted.AvatarURL,
		&inserted.Status,
		&inserted.CreatedAt,
		&inserted.UpdatedAt,
	); err != nil {
		return domain.User{}, fmt.Errorf("create user: %w", err)
	}

	return mapUserRow(inserted), nil
}

// PostgresTokenRepo implements TokenRepository.
type PostgresTokenRepo struct {
	q *sqlc.Queries
}

func NewPostgresTokenRepo(q *sqlc.Queries) *PostgresTokenRepo {
	return &PostgresTokenRepo{q: q}
}

func (r *PostgresTokenRepo) CreateToken(ctx context.Context, token domain.OAuthToken) (domain.OAuthToken, error) {
	refresh := sql.NullString{}
	if token.RefreshToken != "" {
		refresh = sql.NullString{String: token.RefreshToken, Valid: true}
	}
	userID := sql.NullInt64{}
	if token.UserID != 0 {
		userID = sql.NullInt64{Int64: token.UserID, Valid: true}
	}
	row, err := r.q.InsertOAuthToken(ctx, token.ID, token.OrgID, token.ClientID, userID, token.AccessToken, refresh, token.Scopes, token.ExpiresAt)
	if err != nil {
		return domain.OAuthToken{}, fmt.Errorf("insert token: %w", err)
	}
	return mapTokenRow(row), nil
}

func (r *PostgresTokenRepo) GetByRefreshToken(ctx context.Context, orgID int64, token string) (domain.OAuthToken, error) {
	row, err := r.q.GetOAuthTokenByRefresh(ctx, orgID, token)
	if err != nil {
		return domain.OAuthToken{}, fmt.Errorf("get refresh token: %w", err)
	}
	return mapTokenRow(row), nil
}

func (r *PostgresTokenRepo) GetByRefreshTokenValue(ctx context.Context, token string) (domain.OAuthToken, error) {
	row, err := r.q.GetOAuthTokenByRefreshValue(ctx, token)
	if err != nil {
		return domain.OAuthToken{}, fmt.Errorf("get refresh token value: %w", err)
	}
	return mapTokenRow(row), nil
}

func (r *PostgresTokenRepo) GetByAccessToken(ctx context.Context, token string) (domain.OAuthToken, error) {
	row, err := r.q.GetOAuthTokenByAccess(ctx, token)
	if err != nil {
		return domain.OAuthToken{}, fmt.Errorf("get access token: %w", err)
	}
	return mapTokenRow(row), nil
}

func (r *PostgresTokenRepo) RotateRefreshToken(ctx context.Context, tokenID int64, refreshToken string, expiresAt int64) error {
	if err := r.q.RotateRefreshToken(ctx, tokenID, refreshToken, time.Unix(expiresAt, 0)); err != nil {
		return fmt.Errorf("rotate refresh token: %w", err)
	}
	return nil
}

func (r *PostgresTokenRepo) RevokeToken(ctx context.Context, tokenID int64) error {
	if err := r.q.RevokeOAuthToken(ctx, tokenID); err != nil {
		return fmt.Errorf("revoke token: %w", err)
	}
	return nil
}

// PostgresCodeRepo implements CodeRepository.
type PostgresCodeRepo struct {
	q *sqlc.Queries
}

func NewPostgresCodeRepo(q *sqlc.Queries) *PostgresCodeRepo {
	return &PostgresCodeRepo{q: q}
}

func (r *PostgresCodeRepo) CreateCode(ctx context.Context, code domain.OAuthCode) error {
	var challenge sql.NullString
	if code.CodeChallenge != "" {
		challenge = sql.NullString{String: code.CodeChallenge, Valid: true}
	}
	var challengeMethod sql.NullString
	if code.CodeChallengeMethod != "" {
		challengeMethod = sql.NullString{String: code.CodeChallengeMethod, Valid: true}
	}
	if err := r.q.InsertOAuthCode(ctx, code.ID, code.OrgID, code.ClientID, code.UserID, code.Code, code.RedirectURI, challenge, challengeMethod, code.ExpiresAt); err != nil {
		return fmt.Errorf("insert code: %w", err)
	}
	return nil
}

func (r *PostgresCodeRepo) GetCode(ctx context.Context, orgID int64, code string) (domain.OAuthCode, error) {
	row, err := r.q.GetOAuthCode(ctx, orgID, code)
	if err != nil {
		return domain.OAuthCode{}, fmt.Errorf("get code: %w", err)
	}
	return domain.OAuthCode{
		ID:                  row.ID,
		OrgID:               row.TenantID,
		ClientID:            row.ClientID,
		UserID:              row.UserID,
		Code:                row.Code,
		RedirectURI:         row.RedirectURI,
		CodeChallenge:       row.CodeChallenge.String,
		CodeChallengeMethod: row.CodeChallengeMethod.String,
		ExpiresAt:           row.ExpiresAt,
		Revoked:             row.Revoked,
		CreatedAt:           row.CreatedAt,
	}, nil
}

func (r *PostgresCodeRepo) MarkCodeUsed(ctx context.Context, code string) error {
	if err := r.q.RevokeOAuthCode(ctx, code); err != nil {
		return fmt.Errorf("revoke code: %w", err)
	}
	return nil
}

// PostgresKeyRepo implements KeyRepository.
type PostgresKeyRepo struct {
	q *sqlc.Queries
}

func NewPostgresKeyRepo(q *sqlc.Queries) *PostgresKeyRepo {
	return &PostgresKeyRepo{q: q}
}

func (r *PostgresKeyRepo) GetActiveKey(ctx context.Context, orgID int64) (domain.OAuthKey, error) {
	row, err := r.q.GetActiveOAuthKey(ctx, orgID)
	if err != nil {
		return domain.OAuthKey{}, fmt.Errorf("get key: %w", err)
	}
	return mapKeyRow(row), nil
}

func (r *PostgresKeyRepo) CreateKey(ctx context.Context, key domain.OAuthKey) (domain.OAuthKey, error) {
	row, err := r.q.InsertOAuthKey(ctx, key.OrgID, key.KID, key.Secret, key.Algorithm)
	if err != nil {
		return domain.OAuthKey{}, fmt.Errorf("insert key: %w", err)
	}
	mapped := mapKeyRow(row)
	mapped.IsActive = true
	return mapped, nil
}

// PostgresOAuthClientRepo implements OAuthClientRepository.
type PostgresOAuthClientRepo struct {
	db *pgxpool.Pool
}

func NewPostgresOAuthClientRepo(pool *pgxpool.Pool) *PostgresOAuthClientRepo {
	return &PostgresOAuthClientRepo{db: pool}
}

func (r *PostgresOAuthClientRepo) GetClientByID(ctx context.Context, orgID int64, clientID string) (domain.OAuthClient, error) {
	const query = `
SELECT id, tenant_id, app_id, client_id, client_secret, redirect_uris, grants, scopes, token_endpoint_auth_methods, require_consent, created_at
FROM oauth_clients
WHERE tenant_id = $1 AND client_id = $2
LIMIT 1`

	var (
		rowID        int64
		rowTenantID  int64
		rowAppID     sql.NullInt64
		rowClientID  string
		rowSecret    string
		redirectURIs []string
		grants       []string
		scopes       []string
		authMethods  []string
		requireCons  bool
		createdAt    time.Time
	)

	if err := r.db.QueryRow(ctx, query, orgID, clientID).Scan(
		&rowID,
		&rowTenantID,
		&rowAppID,
		&rowClientID,
		&rowSecret,
		&redirectURIs,
		&grants,
		&scopes,
		&authMethods,
		&requireCons,
		&createdAt,
	); err != nil {
		return domain.OAuthClient{}, fmt.Errorf("get oauth client: %w", err)
	}

	var appID *int64
	if rowAppID.Valid {
		val := rowAppID.Int64
		appID = &val
	}

	return domain.OAuthClient{
		ID:                       rowID,
		OrgID:                    rowTenantID,
		AppID:                    appID,
		ClientID:                 rowClientID,
		ClientSecret:             rowSecret,
		RedirectURIs:             append([]string{}, redirectURIs...),
		Grants:                   append([]string{}, grants...),
		Scopes:                   append([]string{}, scopes...),
		TokenEndpointAuthMethods: append([]string{}, authMethods...),
		RequireConsent:           requireCons,
		CreatedAt:                createdAt,
	}, nil
}

func (r *PostgresOAuthClientRepo) UpsertClient(ctx context.Context, client domain.OAuthClient) (domain.OAuthClient, error) {
	const query = `
INSERT INTO oauth_clients (id, tenant_id, app_id, client_id, client_secret, redirect_uris, grants, scopes, token_endpoint_auth_methods, require_consent)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
ON CONFLICT (client_id) DO UPDATE SET
	redirect_uris = EXCLUDED.redirect_uris,
	grants = EXCLUDED.grants,
	scopes = EXCLUDED.scopes,
	token_endpoint_auth_methods = EXCLUDED.token_endpoint_auth_methods,
	require_consent = EXCLUDED.require_consent
RETURNING id, tenant_id, app_id, client_id, client_secret, redirect_uris, grants, scopes, token_endpoint_auth_methods, require_consent, created_at`

	var (
		rowID        int64
		rowTenantID  int64
		rowAppID     sql.NullInt64
		rowClientID  string
		rowSecret    string
		redirectURIs []string
		grants       []string
		scopes       []string
		authMethods  []string
		requireCons  bool
		createdAt    time.Time
	)

	var appID any
	if client.AppID != nil {
		appID = *client.AppID
	}

	if err := r.db.QueryRow(
		ctx,
		query,
		client.ID,
		client.OrgID,
		appID,
		client.ClientID,
		client.ClientSecret,
		client.RedirectURIs,
		client.Grants,
		client.Scopes,
		client.TokenEndpointAuthMethods,
		client.RequireConsent,
	).Scan(
		&rowID,
		&rowTenantID,
		&rowAppID,
		&rowClientID,
		&rowSecret,
		&redirectURIs,
		&grants,
		&scopes,
		&authMethods,
		&requireCons,
		&createdAt,
	); err != nil {
		return domain.OAuthClient{}, fmt.Errorf("upsert oauth client: %w", err)
	}

	var resolvedAppID *int64
	if rowAppID.Valid {
		val := rowAppID.Int64
		resolvedAppID = &val
	}

	return domain.OAuthClient{
		ID:                       rowID,
		OrgID:                    rowTenantID,
		AppID:                    resolvedAppID,
		ClientID:                 rowClientID,
		ClientSecret:             rowSecret,
		RedirectURIs:             append([]string{}, redirectURIs...),
		Grants:                   append([]string{}, grants...),
		Scopes:                   append([]string{}, scopes...),
		TokenEndpointAuthMethods: append([]string{}, authMethods...),
		RequireConsent:           requireCons,
		CreatedAt:                createdAt,
	}, nil
}

// PostgresOAuthAppRepo implements OAuthAppRepository.
type PostgresOAuthAppRepo struct {
	db *pgxpool.Pool
}

func NewPostgresOAuthAppRepo(pool *pgxpool.Pool) *PostgresOAuthAppRepo {
	return &PostgresOAuthAppRepo{db: pool}
}

func (r *PostgresOAuthAppRepo) Create(ctx context.Context, app domain.OAuthApp) (domain.OAuthApp, error) {
	const query = `
INSERT INTO oauth_apps (id, tenant_id, name, app_type, description, icon_url, is_active, is_first_party, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
RETURNING id, tenant_id, name, app_type, description, icon_url, is_active, is_first_party, created_at, updated_at`

	var (
		rowID           int64
		rowTenantID     int64
		rowName         string
		rowType         string
		rowDesc         sql.NullString
		rowIcon         sql.NullString
		rowActive       sql.NullBool
		rowFirstParty   sql.NullBool
		rowCreatedAt    time.Time
		rowUpdatedAt    time.Time
	)

	desc := sql.NullString{}
	if app.Description != "" {
		desc = sql.NullString{String: app.Description, Valid: true}
	}
	icon := sql.NullString{}
	if app.IconURL != "" {
		icon = sql.NullString{String: app.IconURL, Valid: true}
	}

	if err := r.db.QueryRow(ctx, query,
		app.ID,
		app.OrgID,
		app.Name,
		app.Type,
		desc,
		icon,
		app.IsActive,
		app.IsFirstParty,
	).Scan(
		&rowID,
		&rowTenantID,
		&rowName,
		&rowType,
		&rowDesc,
		&rowIcon,
		&rowActive,
		&rowFirstParty,
		&rowCreatedAt,
		&rowUpdatedAt,
	); err != nil {
		return domain.OAuthApp{}, fmt.Errorf("create oauth app: %w", err)
	}

	return domain.OAuthApp{
		ID:           rowID,
		OrgID:        rowTenantID,
		Name:         rowName,
		Type:         rowType,
		Description:  rowDesc.String,
		IconURL:      rowIcon.String,
		IsActive:     rowActive.Bool,
		IsFirstParty: rowFirstParty.Bool,
		CreatedAt:    rowCreatedAt,
		UpdatedAt:    rowUpdatedAt,
	}, nil
}

func (r *PostgresOAuthAppRepo) GetByName(ctx context.Context, orgID int64, name string) (domain.OAuthApp, error) {
	const query = `
SELECT id, tenant_id, name, app_type, description, icon_url, is_active, is_first_party, created_at, updated_at
FROM oauth_apps
WHERE tenant_id = $1 AND name = $2`

	var (
		rowID           int64
		rowTenantID     int64
		rowName         string
		rowType         string
		rowDesc         sql.NullString
		rowIcon         sql.NullString
		rowActive       sql.NullBool
		rowFirstParty   sql.NullBool
		rowCreatedAt    time.Time
		rowUpdatedAt    time.Time
	)

	if err := r.db.QueryRow(ctx, query, orgID, name).Scan(
		&rowID,
		&rowTenantID,
		&rowName,
		&rowType,
		&rowDesc,
		&rowIcon,
		&rowActive,
		&rowFirstParty,
		&rowCreatedAt,
		&rowUpdatedAt,
	); err != nil {
		return domain.OAuthApp{}, fmt.Errorf("get oauth app: %w", err)
	}

	return domain.OAuthApp{
		ID:           rowID,
		OrgID:        rowTenantID,
		Name:         rowName,
		Type:         rowType,
		Description:  rowDesc.String,
		IconURL:      rowIcon.String,
		IsActive:     rowActive.Bool,
		IsFirstParty: rowFirstParty.Bool,
		CreatedAt:    rowCreatedAt,
		UpdatedAt:    rowUpdatedAt,
	}, nil
}

func mapOrgRow(row sqlc.GetTenantRow) domain.Org {
	return domain.Org{
		ID:          row.ID,
		Type:        row.Type,
		Name:        row.Name,
		Code:        row.Code,
		Slug:        row.Slug,
		CountryCode: row.CountryCode,
		Timezone:    row.Timezone,
		IsDefault:   row.IsDefault,
		Status:      row.Status,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}
}

func mapKeyRow(row sqlc.GetActiveOAuthKeyRow) domain.OAuthKey {
	return domain.OAuthKey{
		ID:        row.ID,
		OrgID:     row.TenantID,
		KID:       row.KID,
		Secret:    row.Secret,
		Algorithm: row.Algorithm,
		IsActive:  row.IsActive,
		CreatedAt: row.CreatedAt,
		RotatedAt: nullableTime(row.RotatedAt),
	}
}

func mapTokenRow(row sqlc.InsertOAuthTokenRow) domain.OAuthToken {
	scopes := row.Scopes
	userID := int64(0)
	if row.UserID.Valid {
		userID = row.UserID.Int64
	}
	return domain.OAuthToken{
		ID:           row.ID,
		OrgID:        row.TenantID,
		ClientID:     row.ClientID,
		UserID:       userID,
		AccessToken:  row.AccessToken,
		RefreshToken: row.RefreshToken.String,
		Scopes:       scopes,
		ExpiresAt:    row.ExpiresAt,
		Revoked:      row.Revoked,
		CreatedAt:    row.CreatedAt,
	}
}

func nullableTime(t sql.NullTime) *time.Time {
	if t.Valid {
		return &t.Time
	}
	return nil
}

func mapUserRow(row sqlc.GetUserByEmailRow) domain.User {
	avatar := ""
	if row.AvatarURL.Valid {
		avatar = row.AvatarURL.String
	}
	return domain.User{
		ID:            row.ID,
		OrgID:         row.TenantID,
		Email:         row.Email,
		EmailVerified: row.EmailVerified,
		PasswordHash:  row.PasswordHash,
		Name:          row.Name,
		Phone:         row.Phone,
		PhoneVerified: row.PhoneVerified,
		AvatarURL:     avatar,
		Status:        row.Status,
		CreatedAt:     row.CreatedAt,
		UpdatedAt:     row.UpdatedAt,
	}
}
