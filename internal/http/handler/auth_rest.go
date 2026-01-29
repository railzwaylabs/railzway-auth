package handler

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	domainoauth "github.com/smallbiznis/railzway-auth/internal/domain/oauth"
	"github.com/smallbiznis/railzway-auth/internal/http/middleware"
	"github.com/smallbiznis/railzway-auth/internal/service"
)

func (h *AuthHandler) PasswordLogin(c *gin.Context) {
	orgCtx, ok := middleware.GetOrgContext(c)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "invalid_tenant", "error_description": "Org not resolved."})
		return
	}

	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		ClientID string `json:"client_id"`
		Scope    string `json:"scope"`
		// Optional: when provided, continue OAuth authorize flow using stored state.
		State string `json:"state"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "error_description": "Invalid payload."})
		return
	}
	if strings.TrimSpace(req.Email) == "" || strings.TrimSpace(req.Password) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "error_description": "Email and password are required."})
		return
	}

	authorizeStateID := strings.TrimSpace(req.State)
	authorizeState, err := h.loadAuthorizeState(c, authorizeStateID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "error_description": err.Error()})
		return
	}

	clientID := strings.TrimSpace(req.ClientID)
	if authorizeState != nil {
		if clientID != "" && clientID != authorizeState.ClientID {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "error_description": "client_id does not match authorize state."})
			return
		}
		clientID = authorizeState.ClientID
	}
	if clientID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unauthorized_client", "error_description": "Unknown client_id for org."})
		return
	}

	scope := strings.TrimSpace(req.Scope)

	issuer := fmt.Sprintf("%s://%s", schemeOnly(c.Request), hostOnly(c.Request))
	resp, err := h.Auth.LoginWithPassword(c.Request.Context(), orgCtx.Org.ID, req.Email, req.Password, clientID, scope, issuer)
	if err != nil {
		respondOAuthError(c, err)
		return
	}

	maxAge := 3600
	h.setCookie(c, CookieNameAccessToken, resp.AccessToken, maxAge)
	h.setCookie(c, CookieNameRefreshToken, resp.RefreshToken, maxAge)

	authorizeURL := ""
	if authorizeState != nil {
		authorizeURL = buildAuthorizeURLFromState(authorizeState)
		h.deleteAuthorizeState(c, authorizeStateID)
	}

	c.JSON(http.StatusOK, authResponse{AuthTokensWithUser: resp, AuthorizeURL: authorizeURL})
}

func (h *AuthHandler) PasswordRegister(c *gin.Context) {
	orgCtx, ok := middleware.GetOrgContext(c)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "invalid_tenant", "error_description": "Org not resolved."})
		return
	}

	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Name     string `json:"name"`
		ClientID string `json:"client_id"`
		// Optional: when provided, continue OAuth authorize flow using stored state.
		State string `json:"state"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "error_description": "Invalid payload."})
		return
	}
	if strings.TrimSpace(req.Email) == "" || strings.TrimSpace(req.Password) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "error_description": "Email and password are required."})
		return
	}

	authorizeStateID := strings.TrimSpace(req.State)
	authorizeState, err := h.loadAuthorizeState(c, authorizeStateID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "error_description": err.Error()})
		return
	}

	clientID := strings.TrimSpace(req.ClientID)
	if authorizeState != nil {
		if clientID != "" && clientID != authorizeState.ClientID {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "error_description": "client_id does not match authorize state."})
			return
		}
		clientID = authorizeState.ClientID
	}
	if clientID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unauthorized_client", "error_description": "Unknown client_id for org."})
		return
	}

	issuer := fmt.Sprintf("%s://%s", schemeOnly(c.Request), hostOnly(c.Request))
	resp, err := h.Auth.RegisterWithPassword(c.Request.Context(), orgCtx.Org.ID, req.Email, req.Password, req.Name, clientID, issuer)
	if err != nil {
		respondOAuthError(c, err)
		return
	}

	maxAge := 3600
	h.setCookie(c, CookieNameAccessToken, resp.AccessToken, maxAge)
	h.setCookie(c, CookieNameRefreshToken, resp.RefreshToken, maxAge)

	authorizeURL := ""
	if authorizeState != nil {
		authorizeURL = buildAuthorizeURLFromState(authorizeState)
		h.deleteAuthorizeState(c, authorizeStateID)
	}

	c.JSON(http.StatusOK, authResponse{AuthTokensWithUser: resp, AuthorizeURL: authorizeURL})
}

func (h *AuthHandler) PasswordForgot(c *gin.Context) {
	orgCtx, ok := middleware.GetOrgContext(c)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "invalid_tenant", "error_description": "Org not resolved."})
		return
	}

	var req struct {
		Email string `json:"email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "error_description": "Invalid payload."})
		return
	}
	if strings.TrimSpace(req.Email) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "error_description": "Email is required."})
		return
	}

	if err := h.Auth.ForgotPassword(c.Request.Context(), orgCtx.Org.ID, req.Email); err != nil {
		respondOAuthError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "If the account exists, password reset instructions have been sent."})
}

func (h *AuthHandler) OTPRequest(c *gin.Context) {
	orgCtx, ok := middleware.GetOrgContext(c)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "invalid_tenant", "error_description": "Org not resolved."})
		return
	}

	var req struct {
		Phone   string `json:"phone"`
		Channel string `json:"channel"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "error_description": "Invalid payload."})
		return
	}
	if strings.TrimSpace(req.Phone) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "error_description": "Phone is required."})
		return
	}

	if err := h.Auth.RequestOTP(c.Request.Context(), orgCtx.Org.ID, req.Phone, req.Channel); err != nil {
		respondOAuthError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "OTP request accepted."})
}

func (h *AuthHandler) OTPVerify(c *gin.Context) {
	orgCtx, ok := middleware.GetOrgContext(c)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "invalid_tenant", "error_description": "Org not resolved."})
		return
	}

	var req struct {
		Phone    string `json:"phone"`
		Code     string `json:"code"`
		ClientID string `json:"client_id"`
		Scope    string `json:"scope"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "error_description": "Invalid payload."})
		return
	}
	if strings.TrimSpace(req.Phone) == "" || strings.TrimSpace(req.Code) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "error_description": "Phone and code are required."})
		return
	}

	clientID := strings.TrimSpace(req.ClientID)
	if clientID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unauthorized_client", "error_description": "Unknown client_id for org."})
		return
	}
	scope := strings.TrimSpace(req.Scope)

	issuer := fmt.Sprintf("%s://%s", schemeOnly(c.Request), hostOnly(c.Request))
	resp, err := h.Auth.VerifyOTP(c.Request.Context(), orgCtx.Org.ID, req.Phone, req.Code, clientID, scope, issuer)
	if err != nil {
		respondOAuthError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *AuthHandler) Me(c *gin.Context) {
	orgCtx, ok := middleware.GetOrgContext(c)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "invalid_tenant", "error_description": "Org not resolved."})
		return
	}

	std, ok := middleware.GetStdClaims(c)
	if !ok || std == nil || strings.TrimSpace(std.Subject) == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_token", "error_description": "Missing subject claim."})
		return
	}
	userID, err := strconv.ParseInt(std.Subject, 10, 64)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_token", "error_description": "Invalid subject claim."})
		return
	}

	user, err := h.Auth.GetUserInfo(c.Request.Context(), orgCtx.Org.ID, userID)
	if err != nil {
		respondOAuthError(c, err)
		return
	}

	c.JSON(http.StatusOK, user)
}

func respondOAuthError(c *gin.Context, err error) {
	if oauthErr, ok := err.(*service.OAuthError); ok {
		c.JSON(oauthErr.Status, gin.H{"error": oauthErr.Code, "error_description": oauthErr.Description})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": "server_error", "error_description": err.Error()})
}

func (h *AuthHandler) setCookie(c *gin.Context, name, value string, maxAge int) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(
		name,                      // name
		value,                     // value
		maxAge,                    // maxAge
		"/",                       // path
		c.Request.Host,            // domain
		h.Config.AuthCookieSecure, // secure
		true,                      // httpOnly
	)
}

type authResponse struct {
	service.AuthTokensWithUser
	AuthorizeURL string `json:"authorize_url,omitempty"`
}

func (h *AuthHandler) loadAuthorizeState(c *gin.Context, stateID string) (*domainoauth.AuthorizeState, error) {
	if strings.TrimSpace(stateID) == "" {
		return nil, nil
	}
	if h.AuthorizeStateStore == nil {
		return nil, fmt.Errorf("authorize state store not configured")
	}
	state, err := h.AuthorizeStateStore.GetState(c.Request.Context(), buildAuthorizeStateKey(stateID))
	if err != nil {
		return nil, fmt.Errorf("load authorize state: %w", err)
	}
	if state == nil {
		return nil, fmt.Errorf("authorize state not found")
	}
	orgCtx, ok := middleware.GetOrgContext(c)
	if !ok {
		return nil, fmt.Errorf("org not resolved")
	}
	if state.OrgID != orgCtx.Org.ID {
		return nil, fmt.Errorf("authorize state org mismatch")
	}
	return state, nil
}

func (h *AuthHandler) deleteAuthorizeState(c *gin.Context, stateID string) {
	if strings.TrimSpace(stateID) == "" || h.AuthorizeStateStore == nil {
		return
	}
	if err := h.AuthorizeStateStore.DeleteState(c.Request.Context(), buildAuthorizeStateKey(stateID)); err != nil {
		zap.L().Warn("failed to delete authorize state", zap.Error(err))
	}
}

func buildAuthorizeURLFromState(state *domainoauth.AuthorizeState) string {
	if state == nil {
		return ""
	}
	authorizeURL := &url.URL{Path: "/oauth/authorize"}
	q := authorizeURL.Query()
	q.Set("client_id", state.ClientID)
	responseType := strings.TrimSpace(state.ResponseType)
	if responseType == "" {
		responseType = "code"
	}
	q.Set("response_type", responseType)
	q.Set("redirect_uri", state.RedirectURI)
	if strings.TrimSpace(state.Scope) != "" {
		q.Set("scope", strings.TrimSpace(state.Scope))
	}
	if strings.TrimSpace(state.State) != "" {
		q.Set("state", strings.TrimSpace(state.State))
	}
	if strings.TrimSpace(state.Nonce) != "" {
		q.Set("nonce", strings.TrimSpace(state.Nonce))
	}
	if strings.TrimSpace(state.CodeChallenge) != "" {
		q.Set("code_challenge", strings.TrimSpace(state.CodeChallenge))
	}
	if strings.TrimSpace(state.CodeChallengeMethod) != "" {
		q.Set("code_challenge_method", strings.TrimSpace(state.CodeChallengeMethod))
	}
	authorizeURL.RawQuery = q.Encode()
	return authorizeURL.String()
}
