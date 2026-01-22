package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/smallbiznis/railzway-auth/internal/http/middleware"
	"github.com/smallbiznis/railzway-auth/internal/service"
)

// AdminHandler exposes internal admin endpoints.
type AdminHandler struct {
	Auth *service.AuthService
}

func NewAdminHandler(auth *service.AuthService) *AdminHandler {
	return &AdminHandler{Auth: auth}
}

type upsertOAuthClientRequest struct {
	ExternalOrgID            string   `json:"external_org_id"`
	RedirectURI              string   `json:"redirect_uri"`
	RedirectURIs             []string `json:"redirect_uris"`
	Scopes                   []string `json:"scopes"`
	Grants                   []string `json:"grants"`
	TokenEndpointAuthMethods []string `json:"token_endpoint_auth_methods"`
}

func (h *AdminHandler) UpsertOAuthClient(c *gin.Context) {
	orgCtx, ok := middleware.GetOrgContext(c)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "invalid_tenant", "error_description": "Org not resolved."})
		return
	}

	var req upsertOAuthClientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "error_description": "Invalid payload."})
		return
	}

	externalID := strings.TrimSpace(req.ExternalOrgID)
	if externalID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "error_description": "external_org_id is required."})
		return
	}

	redirectURIs := make([]string, 0, len(req.RedirectURIs)+1)
	for _, uri := range req.RedirectURIs {
		if trimmed := strings.TrimSpace(uri); trimmed != "" {
			redirectURIs = append(redirectURIs, trimmed)
		}
	}
	if trimmed := strings.TrimSpace(req.RedirectURI); trimmed != "" {
		redirectURIs = append(redirectURIs, trimmed)
	}
	if len(redirectURIs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "error_description": "redirect_uri is required."})
		return
	}

	clientID := "railzway-oss-" + externalID
	input := service.OAuthClientInput{
		ClientID:                 clientID,
		RedirectURIs:             redirectURIs,
		Scopes:                   req.Scopes,
		Grants:                   req.Grants,
		TokenEndpointAuthMethods: req.TokenEndpointAuthMethods,
	}

	client, err := h.Auth.UpsertOAuthClient(c.Request.Context(), orgCtx.Org.ID, input)
	if err != nil {
		respondOAuthError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"client_id":                   client.ClientID,
		"client_secret":               client.ClientSecret,
		"redirect_uris":               client.RedirectURIs,
		"scopes":                      client.Scopes,
		"grants":                      client.Grants,
		"token_endpoint_auth_methods": client.TokenEndpointAuthMethods,
	})
}
