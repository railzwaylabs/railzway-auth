package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/smallbiznis/railzway-auth/internal/service"
)

// Admin protects internal admin routes with a shared token.
type Admin struct {
	auth *service.AuthService
}

func NewAdmin(auth *service.AuthService) *Admin {
	return &Admin{auth: auth}
}

// Require validates the admin token header.
func (m *Admin) Require(c *gin.Context) {
	if bearer := readBearerToken(c); bearer != "" {
		if m.auth == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid_token"})
			return
		}
		orgCtx, ok := GetOrgContext(c)
		if !ok || orgCtx == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid_tenant"})
			return
		}
		issuer := fmt.Sprintf("%s://%s", schemeOnly(c.Request), hostOnly(c.Request))
		_, claims, err := m.auth.ValidateToken(c.Request.Context(), orgCtx.Org.ID, bearer, issuer)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid_token"})
			return
		}
		if !hasAdminScope(claims.Scope) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient_scope"})
			return
		}
		c.Next()
		return
	}

	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid_token"})
}

func readBearerToken(c *gin.Context) string {
	if c == nil || c.Request == nil {
		return ""
	}
	header := strings.TrimSpace(c.GetHeader("Authorization"))
	if header == "" {
		return ""
	}
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

func hasAdminScope(scope string) bool {
	if strings.TrimSpace(scope) == "" {
		return false
	}
	for _, value := range strings.Fields(scope) {
		switch strings.ToLower(strings.TrimSpace(value)) {
		case "admin", "admin:*", "admin:oauth_clients":
			return true
		}
	}
	return false
}
