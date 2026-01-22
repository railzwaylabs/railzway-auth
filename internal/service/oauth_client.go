package service

import (
	"context"
	"net/http"
	"strings"

	"github.com/smallbiznis/railzway-auth/internal/domain"
)

// OAuthClientInput describes an OAuth client registration.
type OAuthClientInput struct {
	ClientID                 string
	RedirectURIs             []string
	Scopes                   []string
	Grants                   []string
	TokenEndpointAuthMethods []string
	RequireConsent           bool
}

// UpsertOAuthClient creates or updates an OAuth client for the given org.
func (s *AuthService) UpsertOAuthClient(ctx context.Context, orgID int64, input OAuthClientInput) (domain.OAuthClient, error) {
	if s == nil || s.clients == nil {
		return domain.OAuthClient{}, newOAuthError("server_error", "OAuth client repository unavailable.", http.StatusInternalServerError)
	}

	clientID := strings.TrimSpace(input.ClientID)
	if clientID == "" {
		return domain.OAuthClient{}, newOAuthError("invalid_request", "client_id is required.", http.StatusBadRequest)
	}

	redirectURIs := normalizeList(input.RedirectURIs)
	if len(redirectURIs) == 0 {
		return domain.OAuthClient{}, newOAuthError("invalid_request", "redirect_uris is required.", http.StatusBadRequest)
	}

	scopes := normalizeList(input.Scopes)
	if len(scopes) == 0 {
		scopes = []string{"openid", "profile", "email"}
	}

	grants := normalizeList(input.Grants)
	if len(grants) == 0 {
		grants = []string{"authorization_code", "refresh_token"}
	}

	authMethods := normalizeList(input.TokenEndpointAuthMethods)
	if len(authMethods) == 0 {
		authMethods = []string{"client_secret_post"}
	}

	client := domain.OAuthClient{
		ID:                       s.snowflake.Generate().Int64(),
		OrgID:                    orgID,
		ClientID:                 clientID,
		ClientSecret:             randomString(32),
		RedirectURIs:             redirectURIs,
		Grants:                   grants,
		Scopes:                   scopes,
		TokenEndpointAuthMethods: authMethods,
		RequireConsent:           input.RequireConsent,
	}

	created, err := s.clients.UpsertClient(ctx, client)
	if err != nil {
		return domain.OAuthClient{}, newOAuthError("server_error", "Failed to upsert OAuth client.", http.StatusInternalServerError)
	}

	return created, nil
}

func normalizeList(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	out := make([]string, 0, len(values))
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}
