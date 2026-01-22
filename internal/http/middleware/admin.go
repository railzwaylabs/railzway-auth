package middleware

import (
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/smallbiznis/railzway-auth/internal/config"
)

// Admin protects internal admin routes with a shared token.
type Admin struct {
	cfg config.Config
}

func NewAdmin(cfg config.Config) *Admin {
	return &Admin{cfg: cfg}
}

// Require validates the admin token header.
func (m *Admin) Require(c *gin.Context) {
	expected := strings.TrimSpace(m.cfg.AdminAPIToken)
	if expected == "" {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "admin_token_not_configured"})
		return
	}

	token := strings.TrimSpace(c.GetHeader("X-Admin-Token"))
	if token == "" {
		authHeader := strings.TrimSpace(c.GetHeader("Authorization"))
		if strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
			token = strings.TrimSpace(authHeader[7:])
		}
	}

	if token == "" || subtle.ConstantTimeCompare([]byte(token), []byte(expected)) != 1 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	c.Next()
}
